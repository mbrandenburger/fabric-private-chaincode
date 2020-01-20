/*
 * Copyright IBM Corp. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

#include "auction_cc.h"
#include "auction_json.h"
#include "logging.h"
#include "shim.h"

#include <numeric>
#include <vector>

#define MAX_VALUE_SIZE 1024

#define OK "OK"
#define AUCTION_DRAW "DRAW"
#define AUCTION_NO_BIDS "NO_BIDS"
#define AUCTION_ALREADY_EXISTS "AUCTION_ALREADY_EXISTS"
#define AUCTION_DOES_NOT_EXIST "AUCTION_DOES_NOT_EXIST"
#define AUCTION_ALREADY_CLOSED "AUCTION_ALREADY_CLOSED"
#define AUCTION_STILL_OPEN "AUCTION_STILL_OPEN"

#define INITALIZED_KEY "__enclave_initialized"
#define AUCTION_HOUSE_NAME_KEY "__auction_house_name"

const std::string SEP = ".";
const std::string PREFIX = SEP + "somePrefix" + SEP;

// implements chaincode logic for invoke
int init(
    uint8_t* response, uint32_t max_response_len, uint32_t* actual_response_len, shim_ctx_ptr_t ctx)
{
    // Note: we could have got here not only during instatiation but also due to an upgrade.
    // we do allow this, so don't check status of _initialized here

    LOG_DEBUG("AuctionCC: +++ Executing auction chaincode init+++");
    std::vector<std::string> argss;

    get_string_args(argss, ctx);

    LOG_DEBUG("AuctionCC: Args: %s",
        (argss.size() < 1
                ? "(none)"
                : std::accumulate(std::next(argss.begin()), argss.end(), argss[0],
                      [](std::string a, std::string b) { return (a + std::string(", ") + b); })
                      .c_str()));

    const char* _auction_house_name = argss[0].c_str();
    put_state(
        AUCTION_HOUSE_NAME_KEY, (uint8_t*)_auction_house_name, strlen(_auction_house_name), ctx);

    bool _initialized = true;
    put_state(INITALIZED_KEY, (uint8_t*)&_initialized, sizeof(_initialized), ctx);

    *actual_response_len = 0;
    LOG_DEBUG("AuctionCC: +++ Initialization done +++");
    return 0;
}

// implements chaincode logic for invoke
int invoke(
    uint8_t* response, uint32_t max_response_len, uint32_t* actual_response_len, shim_ctx_ptr_t ctx)
{
    bool _initialized;
    const char* _auction_house_name;
    char _auction_house_name_buf[128];

    uint32_t init_len = -1;
    get_state(INITALIZED_KEY, (uint8_t*)&_initialized, sizeof(_initialized), &init_len, ctx);
    if ((init_len == 0) || !_initialized)
    {
        _initialized = false;
        _auction_house_name = "(uninitialized)";
    }
    else
    {
        uint32_t ahn_len = -1;
        get_state(AUCTION_HOUSE_NAME_KEY, (uint8_t*)_auction_house_name_buf,
            sizeof(_auction_house_name_buf) - 1, &ahn_len, ctx);
        if (ahn_len == 0)
        {
            _auction_house_name = "(uninitialized)";
        }
        else
        {
            _auction_house_name_buf[ahn_len + 1] = '\0';
            _auction_house_name = _auction_house_name_buf;
        }
    }

    LOG_DEBUG(
        "AuctionCC: +++ Executing '%s' auction chaincode invocation +++", _auction_house_name);

    if (!_initialized)
    {
        LOG_ERROR("AuctionCC: Invoke called before initialization");
        *actual_response_len = 0;
        return -1;
    }

    std::string function_name;
    std::vector<std::string> params;
    get_func_and_params(function_name, params, ctx);

    LOG_DEBUG("AuctionCC: Function: %s, Params: %s", function_name.c_str(),
        (params.size() < 1
                ? "(none)"
                : std::accumulate(std::next(params.begin()), params.end(), params[0],
                      [](std::string a, std::string b) { return (a + std::string(", ") + b); })
                      .c_str()));

    std::string auction_name = params[0];
    std::string result;

    if (function_name == "create")
    {
        result = auction_create(auction_name, ctx);
    }
    else if (function_name == "submit")
    {
        int value = std::stoi(params[2]);
        std::string bidder_name = params[1];
        // TODO: eventually replace bidder_name with get_creator_name but for now
        //  in our tests we have only one client, so leave passed bidder_name to
        //  allow for different bidders ...
        char real_bidder_name_msp_id[1024];
        char real_bidder_name_dn[1024];
        get_creator_name(real_bidder_name_msp_id, sizeof(real_bidder_name_msp_id),
            real_bidder_name_dn, sizeof(real_bidder_name_dn), ctx);
        LOG_INFO("AuctionCC: real bidder '(msp_id: %s, dn: %s)' masquerading as '%s'",
            real_bidder_name_msp_id, real_bidder_name_dn, bidder_name.c_str());

        result = auction_submit(auction_name, bidder_name, value, ctx);
    }
    else if (function_name == "close")
    {
        result = auction_close(auction_name, ctx);
    }
    else if (function_name == "eval")
    {
        result = auction_eval(auction_name, ctx);
    }
    else if (function_name == "status")
    {
        result = auction_status(auction_name, ctx);
    }
    else
    {
        // unknown function
        LOG_ERROR("AuctionCC: Received unknown transaction");
        *actual_response_len = 0;
        return -1;
    }

    // check that result fits into response
    int neededSize = result.size();
    if (max_response_len < neededSize)
    {
        // ouch error
        LOG_ERROR("AuctionCC: Response buffer too small");
        *actual_response_len = 0;
        return -1;
    }

    // copy result to response
    memcpy(response, result.c_str(), neededSize);
    *actual_response_len = neededSize;
    LOG_DEBUG("AuctionCC: Response: %s", result.c_str());

    LOG_DEBUG("AuctionCC: +++ Executing done +++");
    return 0;
}

std::string error(std::string msg) {
    status_msg_t status;
    status.rc = 1;
    status.message = msg;
    return marshal_response(&status, "");

}

std::string success(std::string response) {
    status_msg_t status;
    status.rc = 0;
    return marshal_response(&status, response);
}

std::string auction_create(std::string auction_name, shim_ctx_ptr_t ctx)
{
    LOG_DEBUG("Auction CC: Create auction with name %s", auction_name);
    // check if auction already exists
    uint32_t auction_bytes_len = 0;
    uint8_t auction_bytes[MAX_VALUE_SIZE];
    get_state(auction_name.c_str(), auction_bytes, sizeof(auction_bytes), &auction_bytes_len, ctx);

    if (auction_bytes_len > 0)
    {
        // auction already exists
        LOG_DEBUG("AuctionCC: Auction already exists");
        return error(AUCTION_ALREADY_EXISTS);
    }

    // create new auction
    auction_t new_auction;
    new_auction.name = (char*)auction_name.c_str();
    new_auction.is_open = true;

    // convert to json string and store
    std::string json = marshal_auction(&new_auction);
    put_state(auction_name.c_str(), (uint8_t*)json.c_str(), json.size(), ctx);

    return success("{\"auctionId\":1}");
}

std::string auction_submit(
    std::string auction_name, std::string bidder_name, int value, shim_ctx_ptr_t ctx)
{
    LOG_DEBUG("Auction CC: Process bid by %s", bidder_name);
    // check if auction already exists
    uint32_t auction_bytes_len = 0;
    uint8_t auction_bytes[MAX_VALUE_SIZE];
    get_state(auction_name.c_str(), auction_bytes, sizeof(auction_bytes), &auction_bytes_len, ctx);

    if (auction_bytes_len == 0)
    {
        LOG_DEBUG("AuctionCC: Auction does not exist");
        return error(AUCTION_DOES_NOT_EXIST);
    }

    // get auction struct from json
    auction_t the_auction;
    unmarshal_auction(&the_auction, (const char*)auction_bytes, auction_bytes_len);

    if (!the_auction.is_open)
    {
        LOG_DEBUG("AuctionCC: Auction is already closed");
        return error(AUCTION_ALREADY_CLOSED);
    }

    // create composite key "auction_name.bidder_name"
    // if there is already a bid just overwrite it
    std::string new_key(PREFIX + auction_name + SEP + bidder_name + SEP);

    bid_t bid;
    bid.bidder_name = bidder_name;
    bid.value = value;

    // convert to json and store
    std::string json = marshal_bid(&bid);
    put_state(new_key.c_str(), (uint8_t*)json.c_str(), json.size(), ctx);

    return success(OK);
}

std::string auction_close(std::string auction_name, shim_ctx_ptr_t ctx)
{
    LOG_DEBUG("Auction CC: Close auction %s", auction_name);

    // check if auction already exists
    uint32_t auction_bytes_len = 0;
    uint8_t auction_bytes[MAX_VALUE_SIZE];
    get_state(auction_name.c_str(), auction_bytes, sizeof(auction_bytes), &auction_bytes_len, ctx);

    if (auction_bytes_len == 0)
    {
        LOG_DEBUG("AuctionCC: Auction does not exist");
        return error(AUCTION_DOES_NOT_EXIST);
    }

    // get auction struct from json
    auction_t the_auction;
    unmarshal_auction(&the_auction, (const char*)auction_bytes, auction_bytes_len);

    if (!the_auction.is_open)
    {
        LOG_DEBUG("AuctionCC: Auction is already closed");
        return error(AUCTION_ALREADY_CLOSED);
    }

    // close auction
    the_auction.is_open = false;

    // converto to json and store
    std::string json = marshal_auction(&the_auction);
    put_state(auction_name.c_str(), (uint8_t*)json.c_str(), json.size(), ctx);

    return success(OK);
}

std::string auction_eval(std::string auction_name, shim_ctx_ptr_t ctx)
{
    LOG_DEBUG("Auction CC: Evaluate auction %s", auction_name);

    // check if auction already exists
    uint32_t auction_bytes_len = 0;
    uint8_t auction_bytes[MAX_VALUE_SIZE];
    get_state(auction_name.c_str(), auction_bytes, sizeof(auction_bytes), &auction_bytes_len, ctx);

    if (auction_bytes_len == 0)
    {
        LOG_DEBUG("AuctionCC: Auction does not exist");
        return error(AUCTION_DOES_NOT_EXIST);
    }

    // get auction struct from json
    auction_t the_auction;
    unmarshal_auction(&the_auction, (const char*)auction_bytes, auction_bytes_len);

    // check if auction is closed
    if (the_auction.is_open)
    {
        LOG_DEBUG("AuctionCC: Auction is still open");
        return error(AUCTION_STILL_OPEN);
    }

    // get all bids
    std::string bid_composite_key = PREFIX + auction_name + SEP;
    std::map<std::string, std::string> values;
    get_state_by_partial_composite_key(bid_composite_key.c_str(), values, ctx);

    if (values.empty())
    {
        LOG_DEBUG("AuctionCC: No bids");
        return error(AUCTION_NO_BIDS);
    }

    // search highest bid
    bid_t winner;
    int high = -1;
    int draw = 0;

    LOG_DEBUG("AuctionCC: All considered bids:");
    for (auto u : values)
    {
        bid_t b;
        unmarshal_bid(&b, u.second.c_str(), u.second.size());

        LOG_DEBUG("AuctionCC: \t%s value %d", b.bidder_name.c_str(), b.value);
        if (b.value > high)
        {
            draw = 0;
            high = b.value;
            winner = b;
        }
        else if (b.value == high)
        {
            draw = 1;
        }
    }

    if (draw != 1)
    {
        LOG_DEBUG("AuctionCC: Winner is: %s with %d", winner.bidder_name.c_str(), winner.value);
        return success(marshal_bid(&winner));
    }
    else
    {
        LOG_DEBUG("AuctionCC: DRAW");
        return error(AUCTION_DRAW);
    }
}

std::string auction_status(std::string auction_name, shim_ctx_ptr_t ctx)
{
    // check if auction already exists
    uint32_t auction_bytes_len = 0;
    uint8_t auction_bytes[MAX_VALUE_SIZE];
    get_state(auction_name.c_str(), auction_bytes, sizeof(auction_bytes), &auction_bytes_len, ctx);

    if (auction_bytes_len == 0)
    {
        LOG_DEBUG("AuctionCC: Auction does not exist");
        return error(AUCTION_DOES_NOT_EXIST);
    }

    // get auction struct from json
    auction_t the_auction;
    unmarshal_auction(&the_auction, (const char*)auction_bytes, auction_bytes_len);

    return success(marshal_status("clock", 1, the_auction.is_open));
}

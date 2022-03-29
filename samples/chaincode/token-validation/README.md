# Yet another FPC Go Chaincode example

## Building

Make sure you set `$FPC_PATH` to the FPC repository location on your file system.

```bash
make -C  $FPC_PATH/utils/docker pull pull-dev
DOCKER_DEV_OPTIONAL_CMD='make -C samples/chaincode/token-validation' make -C $FPC_PATH/utils/docker run-dev
```

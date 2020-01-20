/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import Login from "@/api/login";

const state = {
  users: []
};

const emptyUser = {
  id: "",
  approle: "",
  color: "",
  avatar: ""
};

const getters = {
  userByName: state => name => {
    return state.users.find(a => a.id === name) || emptyUser;
  },

  userNames(state) {
    return state.users.map(user => user.id);
  },

  avatarByName: (state, getters) => name => {
    return getters.userByName(name).avatar;
  },

  colorByName: (state, getters) => name => {
    return getters.userByName(name).color;
  }
};

const actions = {
  fetchUsers({ commit }) {
    return Login.getRegisteredUsers()
      .then(response => response.data)
      .then(users => commit("setUsers", users));
  }
};

const mutations = {
  setUsers(state, users) {
    state.users = users;
  }
};

export default {
  namespaced: true,
  state,
  getters,
  actions,
  mutations
};

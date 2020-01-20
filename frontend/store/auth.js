import * as jwt from 'jsonwebtoken'
import gql from 'graphql-tag'
// import { oauthConfig } from '~/assets/config'

/**
 * authentication
 */

export const state = () => ({
  token: null,
  user: null,
  loggedIn: false,
  redirectLogin: false
})

export const getters = {
  token: (state) => state.token,
  user: (state) => state.user,
  loggedIn: (state) => state.loggedIn,
  redirectLogin: (state) => state.redirectLogin
}

export const mutations = {
  setRedirectLogin(state, payload) {
    state.redirectLogin = payload
  },
  commitToken(state, payload) {
    state.token = payload
  },
  setUser(state, payload) {
    state.user = payload
  },
  setLoggedIn(state, payload) {
    state.loggedIn = payload
  },
  commitLogout(state) {
    state.token = null
    state.user = null
    state.loggedIn = false
  }
}

export const actions = {
  checkLoggedIn({ state, commit }) {
    let res = true
    try {
      const { exp } = jwt.decode(state.token)
      if (Date.now() >= exp * 1000) {
        res = false
      }
    } catch (err) {
      res = false
    }
    commit('setLoggedIn', res)
    return res
  },
  async loginGoogle({}, payload) {
    return new Promise((resolve, reject) => {
      // use gmail api from script here
    })
  },
  async setToken({ state, commit }, token) {
    return new Promise((resolve, reject) => {
      commit('commitToken', token)
      this.$axios.setToken(state.token)
      this.$apolloHelpers
        .onLogin(state.token)
        .then((res) => {
          resolve('token set')
        })
        .catch((err) => {
          reject(new Error(err))
        })
    })
  },
  async logout({ commit }) {
    return new Promise((resolve, reject) => {
      commit('commitLogout')
      this.$apolloHelpers
        .onLogout()
        .then((res) => {
          resolve('logged out')
        })
        .catch((err) => {
          reject(new Error(err))
        })
    })
  },
  async loginLocal({ commit }, payload) {
    return new Promise((resolve, reject) => {
      this.$axios
        .put('/loginEmailPassword', {
          email: payload.email,
          password: payload.password,
          recaptcha: payload.recaptcha
        })
        .then((res) => {
          if (res.status === 200) {
            if (res.data) {
              if (res.data.token) {
                resolve(res.data.token)
              } else {
                reject(new Error('could not find token data'))
              }
            } else {
              reject(new Error('could not get data'))
            }
          } else {
            reject(new Error(`status code of ${res.status}`))
          }
        })
        .catch((err) => {
          let message = `got error: ${err}`
          if (err.response && err.response.data) {
            message = err.response.data.message
          }
          reject(new Error(message))
        })
    })
  },
  async getUser({ state, commit }) {
    const client = this.app.apolloProvider.defaultClient
    return new Promise((resolve, reject) => {
      if (!state.token) {
        reject(new Error('no token found for user'))
      } else {
        client
          .query({
            query: gql`
              query account {
                account {
                  id
                  email
                  type
                  emailverified
                }
              }
            `,
            variables: {},
            fetchPolicy: 'network-only'
          })
          .then(({ data }) => {
            commit('setUser', data.account)
            resolve('found user account data')
          })
          .catch((err) => {
            console.error(err)
            reject(err)
          })
      }
    })
  }
}

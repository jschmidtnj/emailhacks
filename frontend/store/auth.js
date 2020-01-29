import * as jwt from 'jsonwebtoken'
import gql from 'graphql-tag'
import { getCurrencyAbbreviation } from 'country-currency-map'
// import { oauthConfig } from '~/assets/config'
import { defaultCurrency, defaultCountry } from '~/assets/config'

/**
 * authentication
 */

export const state = () => ({
  token: null,
  user: null,
  country: defaultCountry,
  currency: defaultCurrency,
  exchangeRate: 1,
  loggedIn: false,
  redirectLogin: false
})

export const getters = {
  token: (state) => state.token,
  user: (state) => state.user,
  loggedIn: (state) => state.loggedIn,
  redirectLogin: (state) => state.redirectLogin,
  country: (state) => state.country,
  currency: (state) => state.currency,
  exchangeRate: (state) => state.exchangeRate
}

export const mutations = {
  setCountry(state, payload) {
    state.country = payload
  },
  setCurrency(state, payload) {
    state.currency = payload
  },
  setExchangeRate(state, payload) {
    state.exchangeRate = payload
  },
  setRedirectLogin(state, payload) {
    state.redirectLogin = payload
  },
  commitToken(state, payload) {
    state.token = payload
  },
  setUser(state, payload) {
    state.user = payload
  },
  setPlan(state, payload) {
    if (state.user) {
      state.user.plan = payload
    }
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
  async setCountryGetData({ commit }, countryCode) {
    return new Promise((resolve, reject) => {
      const currencyCode = getCurrencyAbbreviation(countryCode)
      this.$axios
        .get('https://api.exchangeratesapi.io/latest', {
          params: {
            base: defaultCurrency,
            symbols: currencyCode
          },
          baseURL: '',
          headers: null
        })
        .then((res) => {
          if (res.data) {
            const exchangeRate = res.data.rates[currencyCode]
            commit('setCountry', countryCode)
            commit('setCurrency', currencyCode)
            commit('setExchangeRate', exchangeRate)
            resolve('got country and currency')
          } else {
            reject(new Error('cannot find exchange rate data'))
          }
        })
        .catch((err) => {
          reject(err)
        })
    })
  },
  async getCountry({ dispatch }) {
    return new Promise((resolve, reject) => {
      this.$axios
        .get('https://www.cloudflare.com/cdn-cgi/trace', {
          baseURL: '',
          headers: null
        })
        .then((res) => {
          if (res.data) {
            const ipData = res.data.split('\n')
            const countryCode = ipData[8].split('=')[1]
            dispatch('setCountryGetData', countryCode)
              .then((res) => {
                resolve(res)
              })
              .catch((err) => {
                reject(err)
              })
          } else {
            reject(new Error('cannot get country data'))
          }
        })
        .catch((err) => {
          reject(err)
        })
    })
  },
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
                  emailverified
                  type
                  categories {
                    name
                    color
                  }
                  tags {
                    name
                    color
                  }
                  plan
                  purchases
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

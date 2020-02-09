import gql from 'graphql-tag'

export const state = () => ({
  plan: null,
  products: [],
  productOptions: [],
  countryOptions: [],
  currencyOptions: []
})

export const getters = {
  plan: (state) => state.plan,
  products: (state) => state.products,
  productOptions: (state) => state.productOptions,
  countryOptions: (state) => state.countryOptions,
  currencyOptions: (state) => state.currencyOptions,
  total: (state) => {
    /* eslint-disable */
    return state.plan
      ? state.productOptions[state.plan.productIndex].plans[
          state.plan.planIndex
        ].amount
      : 0 + state.products.length > 0
      ? state.products
          .map((productIndex) => state.productOptions[productIndex])
          .reduce((acc, curr) => (acc += curr.plans[0].amount))
      : 0
    /* eslint-enable */
  }
}

export const mutations = {
  reset(state) {
    state.plan = null
    state.products = []
  },
  setPlan(state, payload) {
    state.plan = payload
  },
  setProductOptions(state, payload) {
    state.productOptions = payload
  },
  setCountryOptions(state, payload) {
    state.countryOptions = payload
  },
  setCurrencyOptions(state, payload) {
    state.currencyOptions = payload
  },
  addCurrencyOption(state, payload) {
    if (!state.currencyOptions.includes(payload))
      state.currencyOptions.push(payload)
  },
  removeCurrencyOption(state, payload) {
    const index = state.currencyOptions.indexOf(payload)
    if (index !== -1) state.currencyOptions.splice(index, 1)
  },
  addProduct(state, payload) {
    if (!state.products.includes(payload)) {
      state.products.push(payload)
    }
  },
  removeProduct(state, payload) {
    state.products.splice(payload, 1)
  }
}

export const actions = {
  addPlan({ commit, state }, selection) {
    if (
      selection.productIndex < 0 ||
      selection.productIndex > state.productOptions.length
    ) {
      return new Error('invalid index found')
    }
    if (
      !state.productOptions[selection.productIndex].plans ||
      state.productOptions[selection.productIndex].plans.length === 0
    ) {
      return new Error('cannot find any plans for given product')
    }
    if (
      selection.planIndex < 0 ||
      selection.planIndex >=
        state.productOptions[selection.productIndex].plans.length
    ) {
      return new Error('plan index is invalid')
    }
    const planType =
      state.productOptions[selection.productIndex].plans[selection.planIndex]
        .type
    if (planType === 'once') {
      // single purchase
      commit('addProduct', selection.productIndex)
    } else {
      commit('setPlan', selection)
    }
  },
  async getProductOptions({ commit }) {
    const client = this.app.apolloProvider.defaultClient
    return new Promise((resolve, reject) => {
      client
        .query({
          query: gql`
            query products {
              products {
                id
                name
                plans {
                  amount
                  interval
                }
                maxprojects
                maxforms
                maxstorage
              }
            }
          `,
          variables: {},
          fetchPolicy: 'network-only'
        })
        .then(({ data }) => {
          commit('setProductOptions', data.products)
          resolve('found product options')
        })
        .catch((err) => {
          console.error(err)
          reject(err)
        })
    })
  },
  async getCountryOptions({ commit }) {
    const client = this.app.apolloProvider.defaultClient
    return new Promise((resolve, reject) => {
      client
        .query({
          query: gql`
            query countries {
              countries
            }
          `,
          variables: {}
        })
        .then(({ data }) => {
          commit('setCountryOptions', data.countries)
          resolve('found country options')
        })
        .catch((err) => {
          console.error(err)
          reject(err)
        })
    })
  },
  async getCurrencyOptions({ commit }, useCache) {
    const client = this.app.apolloProvider.defaultClient
    return new Promise((resolve, reject) => {
      client
        .query({
          query: gql`
            query currencies($useCache: Boolean!) {
              currencies(useCache: $useCache)
            }
          `,
          variables: { useCache },
          fetchPolicy: useCache ? 'cache-first' : 'network-only'
        })
        .then(({ data }) => {
          commit('setCurrencyOptions', data.currencies)
          resolve('found currency options')
        })
        .catch((err) => {
          console.error(err)
          reject(err)
        })
    })
  }
}

import gql from 'graphql-tag'

const validProductTypes = ['plan', 'product']

export const state = () => ({
  plan: null,
  products: [],
  options: []
})

export const getters = {
  plan: (state) => state.plan,
  products: (state) => state.products,
  options: (state) => state.options,
  total: (state) => {
    return (
      state.options[state.plan.productIndex].plans[state.plan.planIndex]
        .amount +
      state.products
        .map((productIndex) => state.options[productIndex])
        .reduce((acc, curr) => (acc += curr.plans[0].amount))
    )
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
  setOptions(state, payload) {
    state.options = payload
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
    if (!validProductTypes.includes(selection)) {
      return new Error('invalid product type found')
    }
    if (!selection.productIndex) {
      return new Error('cannot find index')
    }
    if (
      selection.productIndex < 0 ||
      selection.productIndex > state.options.length
    ) {
      return new Error('invalid index found')
    }
    if (
      !selection.options[selection.productIndex].plans ||
      selection.options[selection.productIndex].plans.length === 0
    ) {
      return new Error('cannot find any plans for given product')
    }
    if (selection.type === validProductTypes[1]) {
      // single purchase
      if (state.options[selection.productIndex] !== 'once') {
        return new Error('current selection is not a one-time purchase')
      }
      commit('addProduct', selection.productIndex)
    } else {
      if (
        !selection.planIndex ||
        selection.planIndex < 0 ||
        selection.planIndex >=
          selection.options[selection.productIndex].plans.length
      ) {
        return new Error('plan index is invalid')
      }
      commit('setPlan', selection)
    }
  },
  async getOptions({ commit }) {
    const client = this.app.apolloProvider.defaultClient
    return new Promise((resolve, reject) => {
      client
        .query({
          query: gql`
            query products {
              products {
                id
                name
                plans
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
          commit('setOptions', data.products)
          resolve('found options')
        })
        .catch((err) => {
          console.error(err)
          reject(err)
        })
    })
  }
}

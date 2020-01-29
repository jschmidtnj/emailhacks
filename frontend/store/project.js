export const state = () => ({
  project: null
})

export const getters = {
  project: (state) => state.project
}

export const mutations = {
  setProject(state, payload) {
    state.project = payload
  }
}

export const actions = {}

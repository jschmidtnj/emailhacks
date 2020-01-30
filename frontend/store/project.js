import gql from 'graphql-tag'
export const state = () => ({
  projectId: null,
  projectName: null
})

export const getters = {
  projectId: (state) => state.projectId,
  projectName: (state) => state.projectName
}

export const mutations = {
  setProjectId(state, payload) {
    state.projectId = payload
  },
  setProjectName(state, payload) {
    state.projectName = payload
  }
}

export const actions = {
  async getProjectName({ store, commit }) {
    const client = this.app.apolloProvider.defaultClient
    return new Promise((resolve, reject) => {
      client
        .query({
          query: gql`
            query project($id: String!) {
              project(id: $id) {
                name
              }
            }
          `,
          variables: { id: store.projectId },
          fetchPolicy: 'network-only'
        })
        .then(({ data }) => {
          commit('setProjectName', data.project.name)
          resolve('got project')
        })
        .catch((err) => {
          commit('setProjectName', null)
          commit('setProjectId', null)
          reject(err)
        })
    })
  }
}

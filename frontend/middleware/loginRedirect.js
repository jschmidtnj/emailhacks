const loginRedirect = ({ store, redirect, query }) => {
  return new Promise((resolve, reject) => {
    if (!store.state.auth) {
      resolve()
    } else {
      store
        .dispatch('auth/checkLoggedIn')
        .then((loggedin) => {
          if (!loggedin) {
            resolve()
          } else if (!store.state.auth.user) {
            store
              .dispatch('auth/getUser')
              .then((res) => {
                if (!query.redirect_uri) {
                  redirect('/profile')
                } else {
                  resolve()
                }
              })
              .catch((err) => {
                resolve()
              })
          } else if (!query.redirect_uri) {
            redirect('/profile')
          } else {
            resolve()
          }
        })
        .catch((err) => {
          resolve()
        })
    }
  })
}

export default loginRedirect

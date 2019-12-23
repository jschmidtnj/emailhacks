const admin = ({ store, redirect }) => {
  return new Promise((resolve, reject) => {
    if (!store.state.auth) {
      redirect('/login')
    } else {
      store
        .dispatch('auth/checkLoggedIn')
        .then((loggedIn) => {
          if (!loggedIn) {
            redirect('/login')
          } else if (!store.state.auth.user) {
            store
              .dispatch('auth/getUser')
              .then((res) => {
                if (store.state.auth.user.type === 'admin') {
                  resolve()
                } else {
                  redirect('/login')
                }
              })
              .catch((err) => {
                redirect('/login')
              })
          } else if (store.state.auth.user.type === 'admin') {
            resolve()
          } else {
            redirect('/login')
          }
        })
        .catch((err) => {
          redirect('/login')
        })
    }
  })
}

export default admin

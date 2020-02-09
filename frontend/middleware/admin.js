import { adminTypes } from '~/assets/config'
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
                if (adminTypes.includes(store.state.auth.user.type)) {
                  resolve()
                } else {
                  redirect('/login')
                }
              })
              .catch((err) => {
                console.error(err)
                redirect('/login')
              })
          } else if (adminTypes.includes(store.state.auth.user.type)) {
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

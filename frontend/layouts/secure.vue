<template>
  <div>
    <sidebar />
    <div class="main-wrapper main-wrapper-sidebar">
      <navbar />
      <nuxt class="content" />
      <main-footer />
    </div>
  </div>
</template>

<script lang="js">
import Vue from 'vue'
import Navbar from '~/components/Navbar.vue'
import Sidebar from '~/components/Sidebar.vue'
import MainFooter from '~/components/Footer.vue'
import { checkLoggedInInterval, adminTypes } from '~/assets/config'
export default Vue.extend({
  name: 'Secure',
  // @ts-ignore
  middleware: 'auth',
  components: {
    Navbar,
    Sidebar,
    MainFooter
  },
  data() {
    return {
      interval: null
    }
  },
  computed: {
    admin() {
      return (
        this.$store.state.auth &&
        this.$store.state.auth.user &&
        adminTypes.includes(this.store.state.auth.user.type)
      )
    }
  },
  // @ts-ignore
  head() {
    // @ts-ignore
    const seo = JSON.parse(process.env.seoconfig)
    const links = []
    const meta = []
    if (seo) {
      const canonical = `${seo.url}/${this.$route.path}`
      links.push({
        rel: 'canonical',
        href: canonical
      })
      meta.push({
        property: 'og:url',
        content: canonical
      })
    }
    return {
      links,
      meta
    }
  },
  mounted() {
    this.interval = setInterval(() => {
      this.$store
        .dispatch('auth/checkLoggedIn')
        .then((loggedIn) => {
          if (!loggedIn) {
            this.$store.dispatch('auth/logout').then(() => {
              this.$router.push({
                path: '/login'
              })
            }).catch(err => {
              this.$toasted.global.error({
                message: err
              })
            })
          }
        })
        .catch((err) => {
          this.$store.dispatch('auth/logout').then(() => {
            this.$router.push({
              path: '/login'
            })
          }).catch(err => {
            this.$toasted.global.error({
              message: err
            })
          })
        })
    }, checkLoggedInInterval)
  },
  beforeDestroy() {
    if (this.interval) {
      clearInterval(this.interval)
    }
  }
})
</script>

<style lang="scss"></style>

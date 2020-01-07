<template>
  <div>
    <div v-if="loggedIn">
      <sidebar />
      <div class="main-wrapper main-wrapper-sidebar">
        <navbar />
        <nuxt class="content" />
        <main-footer />
      </div>
    </div>
    <div v-else class="main-wrapper">
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
import { checkLoggedInInterval } from '~/assets/config'
export default Vue.extend({
  name: 'Default',
  components: {
    Navbar,
    Sidebar,
    MainFooter
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
  data() {
    return {
      interval: null
    }
  },
  computed: {
    loggedIn() {
      return this.$store.state.auth && this.$store.state.auth.loggedIn
    }
  },
  mounted() {
    this.interval = setInterval(() => {
      this.$store
        .dispatch('auth/checkLoggedIn')
        .then((loggedIn) => {
          if (!loggedIn) {
            this.$store.dispatch('auth/logout').then(() => {
              console.log('logged out')
            }).catch(err => {
              this.$toasted.global.error({
                message: err
              })
            })
          }
        })
        .catch((err) => {
          this.$store.dispatch('auth/logout').then(() => {
            console.log('logged out')
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

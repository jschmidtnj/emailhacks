<template>
  <div class="main-wrapper">
    <navbar />
    <nuxt class="content" />
    <main-footer />
  </div>
</template>

<script lang="js">
import Vue from 'vue'
import Navbar from '~/components/Navbar.vue'
import MainFooter from '~/components/Footer.vue'
import { checkLoggedInInterval } from '~/assets/config'
export default Vue.extend({
  name: 'Admin',
  // @ts-ignore
  middleware: 'admin',
  components: {
    Navbar,
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
  mounted() {
    this.interval = setInterval(() => {
      this.$store
        .dispatch('auth/checkLoggedIn')
        .then((loggedIn) => {
          if (!loggedIn) {
            this.$store.commit('auth/logout')
            this.$router.push({
              path: '/login'
            })
          }
        })
        .catch((err) => {
          this.$store.commit('auth/logout')
          this.$router.push({
            path: '/login'
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

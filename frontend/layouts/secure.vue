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
    const i18nSeo = this.$nuxtI18nSeo()
    const links = [...i18nSeo.link]
    const meta = [...i18nSeo.meta]
    const htmlAttrs = {
      ...i18nSeo.htmlAttrs
    }
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
      meta,
      htmlAttrs
    }
  },
  mounted() {
    this.checkLoggedIn()
    this.interval = setInterval(() => {
      this.checkLoggedIn()
    }, checkLoggedInInterval)
  },
  beforeDestroy() {
    if (this.interval) {
      clearInterval(this.interval)
    }
  },
  methods: {
    checkLoggedIn() {
      this.$store
        .dispatch('auth/checkLoggedIn')
        .then((loggedIn) => {
          if (!loggedIn) {
            this.$store.dispatch('auth/logout').then(() => {
              this.$router.push({
                path: '/login'
              })
            }).catch(err => {
              this.$bvToast.toast(err, {
                variant: 'danger',
                title: 'Error'
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
            this.$bvToast.toast(err, {
              variant: 'danger',
              title: 'Error'
            })
          })
        })
    }
  }
})
</script>

<style lang="scss"></style>

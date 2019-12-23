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
export default Vue.extend({
  name: 'Secure',
  // @ts-ignore
  middleware: 'auth',
  components: {
    Navbar,
    MainFooter
  },
  computed: {
    admin() {
      return (
        this.$store.state.auth &&
        this.$store.state.auth.user &&
        this.$store.state.auth.user.type === 'admin'
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
  }
})
</script>

<style lang="scss"></style>

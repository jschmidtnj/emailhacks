<template>
  <div>
    <p>Admin Page</p>
    <b-btn @click="logout" block>
      Logout
    </b-btn>
  </div>
</template>

<script lang="js">
import Vue from 'vue'
// @ts-ignore
const seo = JSON.parse(process.env.seoconfig)
export default Vue.extend({
  // @ts-ignore
  layout: 'admin',
  // @ts-ignore
  head() {
    const title = 'Admin Home'
    const description = 'main admin dashboard'
    const image = `${seo.url}/icon.png`
    return {
      title,
      meta: [
        { property: 'og:title', content: title },
        { property: 'og:description', content: description },
        {
          property: 'og:image',
          content: image
        },
        { name: 'twitter:title', content: title },
        {
          name: 'twitter:description',
          content: description
        },
        {
          name: 'twitter:image',
          content: image
        },
        { hid: 'description', name: 'description', content: description }
      ]
    }
  },
  methods: {
    logout(evt) {
      evt.preventDefault()
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
  }
})
</script>

<style lang="scss"></style>

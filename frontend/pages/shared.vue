<template>
  <b-container class="mt-4">
    <form-list />
  </b-container>
</template>

<script lang="js">
import Vue from 'vue'
import FormList from '~/components/form/SharedFormList.vue'
// @ts-ignore
const seo = JSON.parse(process.env.seoconfig)
// contains all shared forms
export default Vue.extend({
  name: 'Forms',
  layout: 'secure',
  components: {
    FormList
  },
  data() {
    return {}
  },
  // @ts-ignore
  head() {
    const title = 'Search Forms'
    const description = 'search for forms'
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
  }
})
</script>

<style lang="scss"></style>

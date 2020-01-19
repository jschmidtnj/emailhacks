<template>
  <b-container class="mt-4">
    <response-list />
  </b-container>
</template>

<script lang="js">
import Vue from 'vue'
import ResponseList from '~/components/response/ResponseList.vue'
// @ts-ignore
const seo = JSON.parse(process.env.seoconfig)
export default Vue.extend({
  name: 'AllResponses',
  layout: 'secure',
  components: {
    ResponseList
  },
  data() {
    return {}
  },
  // @ts-ignore
  head() {
    const title = 'View Responses'
    const description = 'responses for forms'
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

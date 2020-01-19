<template>
  <b-container class="mt-4">
    <response-list
      v-if="projectId && formId && responseId"
      :form-id="formId"
      :project-id="projectId"
    />
  </b-container>
</template>

<script lang="js">
import Vue from 'vue'
import ResponseList from '~/components/response/ResponseList.vue'
// @ts-ignore
const seo = JSON.parse(process.env.seoconfig)
export default Vue.extend({
  name: 'Responses',
  layout: 'secure',
  components: {
    ResponseList
  },
  data() {
    return {
      projectId: null,
      formId: null
    }
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
  },
  mounted() {
    if (this.$route.params && this.$route.params.projectId && this.$route.params.formId) {
      this.projectId = this.$route.params.projectId
      this.formId = this.$route.params.formId
    } else {
      this.$nuxt.error({
        statusCode: 404,
        message: 'could not find form id or project id or response id'
      })
    }
  }
})
</script>

<style lang="scss"></style>

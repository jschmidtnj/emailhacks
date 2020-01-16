<template>
  <b-container class="mt-4">
    <create
      v-if="projectId && formId"
      :form-id="formId"
      :project-id="projectId"
      :get-initial-data="true"
    />
  </b-container>
</template>

<script lang="js">
import Vue from 'vue'
import Create from '~/components/secure/form/Create.vue'
const seo = JSON.parse(process.env.seoconfig)
export default Vue.extend({
  name: 'EditForm',
  layout: 'secure',
  components: {
    Create
  },
  head() {
    const title = 'Edit Form'
    const description = 'edit a form'
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
  data() {
    return {
      projectId: null,
      formId: null
    }
  },
  mounted() {
    if (this.$route.params && this.$route.params.projectId && this.$route.params.formId) {
      this.projectId = this.$route.params.projectId
      this.formId = this.$route.params.formId
    } else {
      this.$nuxt.error({
        statusCode: 404,
        message: 'could not find form id or project id'
      })
    }
  }
})
</script>

<style lang="scss"></style>

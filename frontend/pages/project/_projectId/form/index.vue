<template>
  <b-container class="mt-4">
    <create
      v-if="formId && projectId"
      :form-id="formId"
      :project-id="projectId"
    />
  </b-container>
</template>

<script lang="js">
import Vue from 'vue'
import gql from 'graphql-tag'
import Create from '~/components/form/Create.vue'
import { defaultItemName } from '~/assets/config'
const seo = JSON.parse(process.env.seoconfig)
export default Vue.extend({
  name: 'NewForm',
  layout: 'secure',
  components: {
    Create
  },
  head() {
    const title = 'New Form'
    const description = 'create a form'
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
    if (this.$route.params && this.$route.params.projectId) {
      this.projectId = this.$route.params.projectId
      this.$apollo.mutate({mutation: gql`
        mutation addForm($project: String!, $name: String!, $items: [FormItemInput!]!, $multiple: Boolean!, $files: [FileInput!]!, $tags: [String!]!, $categories: [String!]!)
        {addForm(project: $project, name: $name, items: $items, multiple: $multiple, files: $files, tags: $tags, categories: $categories){id} }
        `, variables: {project: this.projectId, name: defaultItemName, items: [], multiple: false, files: [], categories: [], tags: []}})
        .then(({ data }) => {
          this.formId = data.addForm.id
          history.replaceState({}, null, `/project/${this.projectId}/form/${this.formId}/edit`)
        }).catch(err => {
          console.error(err)
          this.$toasted.global.error({
            message: `found error: ${err.message}`
          })
        })
    } else {
      this.$nuxt.error({
        statusCode: 404,
        message: 'could not find project id'
      })
    }
  }
})
</script>

<style lang="scss"></style>

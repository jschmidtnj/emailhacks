<template>
  <b-container class="mt-4">
    <create v-if="formId" :form-id="formId" />
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
      formId: null
    }
  },
  mounted() {
    this.$apollo.mutate({mutation: gql`
      mutation addForm($project: String!, $name: String!, $items: [FormItemInput!]!, $multiple: Boolean!, $files: [FileInput!]!, $tags: [String!]!, $categories: [String!]!)
      {addForm(project: $project, name: $name, items: $items, multiple: $multiple, files: $files, tags: $tags, categories: $categories){id} }
      `, variables: {project: this.$store.state.project.projectId, name: defaultItemName, items: [], multiple: false, files: [], categories: [], tags: []}})
      .then(({ data }) => {
        this.formId = data.addForm.id
        history.replaceState({}, null, `form/${this.formId}/edit`)
      }).catch(err => {
        console.error(err)
        this.$bvToast.toast(`found error: ${err.message}`, {
          variant: 'danger',
          title: 'Error'
        })
      })
  }
})
</script>

<style lang="scss"></style>

<template>
  <b-container v-if="formId" class="mt-4">
    <response-list
      v-if="userAccess"
      :form-id="formId"
      :edit-access="userAccess.type === 'edit'"
    />
  </b-container>
</template>

<script lang="js">
import Vue from 'vue'
import gql from 'graphql-tag'
import ResponseList from '~/components/response/ResponseList.vue'
// advanced search responses
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
      formId: null,
      formData: null
    }
  },
  computed: {
    userAccess() {
      return this.formData ? this.formData.access.find(elem => elem.id === this.$store.state.auth.user.id) : null
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
    if (this.$route.params && this.$route.params.formId) {
      this.formId = this.$route.params.formId
      this.$apollo.query({
        query: gql`
          query form($id: String!){
            form(id: $id) {
              access
            }
          }`,
          variables: {id: this.formId},
          fetchPolicy: 'network-only'
        })
        .then(({ data }) => {
          this.formData = data.form
        }).catch(err => {
          console.error(err.message)
          this.$bvToast.toast(`found error: ${err.message}`, {
            variant: 'danger',
            title: 'Error'
          })
        })
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

<template>
  <b-container v-if="formId" class="mt-4">
    <b-row>
      <b-col>
        <response-list
          v-if="userAccess"
          :form-id="formId"
          :edit-access="userAccess.type === 'edit'"
        />
      </b-col>
      <b-col>
        <nuxt-link
          :to="`/form/${formId}/view`"
          class="btn btn-primary btn-sm no-underline mt-4"
        >
          view
        </nuxt-link>
        <b-container v-if="!responseId">
          <vc-donut
            :size="200"
            :thickness="30"
            :sections="chartSections"
            :total="100"
            :start-angle="0"
            background="white"
            foreground="grey"
            unit="px"
            has-legend
            legend-placement="top"
          >
            {{ responses }} responses / {{ views }} views
          </vc-donut>
        </b-container>
        <b-container v-else>
          <view-content
            :form-id="formId"
            :response-id="responseId"
            :form-data="formData"
          />
        </b-container>
      </b-col>
    </b-row>
  </b-container>
</template>

<script lang="js">
import Vue from 'vue'
import gql from 'graphql-tag'
import ResponseList from '~/components/response/Responses.vue'
import ViewContent from '~/components/form/View.vue'
// @ts-ignore
const seo = JSON.parse(process.env.seoconfig)
export default Vue.extend({
  name: 'Responses',
  layout: 'secure',
  components: {
    ResponseList,
    ViewContent
  },
  data() {
    return {
      formId: null,
      responseId: null,
      formData: null
    }
  },
  computed: {
    chartTotal() {
      return this.chartSections.reduce((acc, curr) => acc + curr.value)
    },
    chartSections() {
      return [
        { label: 'views', value: this.views, color: 'blue' },
        { label: 'responses', value: this.responses, color: 'green' },
      ]
    },
    views() {
      return this.formData ? this.formData.views : 0
    },
    responses() {
      return this.formData ? this.formData.responses : 0
    },
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
          query form($id: String!, $accessToken: String){
            form(id: $id, editAccessToken: false, accessToken: $accessToken) {
              views
              responses
              name
              items {
                question
                type
                options
                text
                required
                files
              }
              access {
                id
                type
              }
              multiple
              files {
                id
                name
                type
                originalSrc
              }
              ${!this.currentResponseId ? 'updatesAccessToken' : ''}
            }
          }`,
          variables: {id: this.formId, accessToken: this.accessToken},
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

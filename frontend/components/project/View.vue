<template>
  <b-container>
    <b-container>
      <b-form @submit="updateProject">
        <b-form-group>
          <b-input-group>
            <b-form-input v-model="name" />
          </b-input-group>
        </b-form-group>
      </b-form>
    </b-container>
    <form-list :project-id="projectId" />
    <nuxt-link
      :to="`/project/${projectId}/form`"
      class="btn btn-primary btn-sm no-underline mt-4"
    >
      Create New Form
    </nuxt-link>
  </b-container>
</template>

<script lang="js">
import Vue from 'vue'
import gql from 'graphql-tag'
import FormList from '~/components/form/FormList.vue'
import { defaultItemName, noneAccessType } from '~/assets/config'
// @ts-ignore
const seo = JSON.parse(process.env.seoconfig)
export default Vue.extend({
  name: 'ViewProject',
  components: {
    FormList
  },
  props: {
    projectId: {
      type: String,
      default: null
    },
    getInitialData: {
      type: Boolean,
      default: true
    }
  },
  data() {
    return {
      name: '',
      isPublic: false
    }
  },
  // @ts-ignore
  head() {
    const title = 'Search Forms'
    const description = 'search for forms, by name, views, etc'
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
    if (this.getInitialData) {
      this.getProject()
    } else {
      this.name = defaultItemName
    }
  },
  methods: {
    getProject() {
      this.$apollo.query({query: gql`
        query project($id: String!) {
          project(id: $id) {
            name
            public
          }
        }
        `, variables: {id: this.projectId}})
        .then(({ data }) => {
          this.name = data.project.name
          this.isPublic = data.project.public === noneAccessType
          this.$store.commit('auth/setRedirectLogin', this.isPublic)
        }).catch(err => {
          console.error(err)
          this.$toasted.global.error({
            message: `found error: ${err.message}`
          })
        })
    },
    updateProject(evt) {
      evt.preventDefault()
      this.$apollo.mutate({mutation: gql`
        mutation updateProject($id: String!, $name: String!){updateProject(id: $id, name: $name){id} }
        `, variables: {id: this.projectId, name: this.name}})
        .then(({ data }) => {
          console.log('updated!')
        }).catch(err => {
          console.error(err)
          this.$toasted.global.error({
            message: `found error: ${err.message}`
          })
        })
    }
  }
})
</script>

<style lang="scss"></style>

<template>
  <b-container>
    <b-container>
      <b-form @submit="updateProject">
        <b-form-group>
          <b-input-group>
            <b-form-input v-model="name"></b-form-input>
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
import { format } from 'date-fns'
import FormList from '~/components/secure/form/FormList.vue'
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
      name: ''
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
      this.name = 'Untitled'
    }
  },
  methods: {
    getProject() {
      this.$axios.get('/graphql', {
        params: {
          query: `{project(id:"${this.projectId}"){name}}`
        }
      })
      .then(res => {
        if (res.status === 200) {
          if (res.data) {
            if (res.data.data && res.data.data.project) {
              this.name = decodeURIComponent(res.data.data.project.name)
            } else if (res.data.errors) {
              console.error(res.data.errors[0])
              this.$toasted.global.error({
                message: `found errors: ${JSON.stringify(res.data.errors)}`
              })
            } else {
              this.$toasted.global.error({
                message: 'could not find data or errors'
              })
            }
          } else {
            this.$toasted.global.error({
              message: 'could not get data'
            })
          }
        } else {
          this.$toasted.global.error({
            message: `status code of ${res.status}`
          })
        }
      })
      .catch(err => {
        console.error(err)
        this.$toasted.global.error({
          message: err
        })
      })
    },
    updateProject(evt) {
      evt.preventDefault()
      this.$axios
        .post('/graphql', {
          query: `mutation{updateProject(id:"${encodeURIComponent(
            this.projectId
          )}",name:"${encodeURIComponent(
            this.name
          )}"){id}}`
        })
        .then(res => {
          if (res.status === 200) {
            if (res.data) {
              if (res.data.data && res.data.data.updateProject) {
                console.log('updated!')
              } else if (res.data.errors) {
                console.log(res.data.errors[0])
                this.$toasted.global.error({
                  message: `found errors: ${JSON.stringify(res.data.errors)}`
                })
              } else {
                this.$toasted.global.error({
                  message: 'could not find data or errors'
                })
              }
            } else {
              this.$toasted.global.error({
                message: 'could not get data'
              })
            }
          } else {
            this.$toasted.global.error({
              message: `status code of ${res.status}`
            })
          }
        })
        .catch(err => {
          this.$toasted.global.error({
            message: err
          })
        })
    },
    formatDate(dateUTC, formatStr) {
      return format(dateUTC, formatStr)
    }
  }
})
</script>

<style lang="scss"></style>

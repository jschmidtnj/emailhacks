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
      <b-row>
        <b-col>
          <form-list />
        </b-col>
        <b-col>
          Project analytics go here...
        </b-col>
      </b-row>
    </b-container>
    <b-button
      @click="newForm"
      pill
      variant="primary"
      class="new-form-button shadow-lg"
    >
      <client-only>
        <font-awesome-icon size="3x" icon="plus" />
      </client-only>
    </b-button>
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
    newForm(evt) {
      evt.preventDefault()
      this.$router.push({
        path: '/form'
      })
    },
    getProject() {
      this.$apollo.query({
        query: gql`
          query project($id: String!) {
            project(id: $id) {
              name
              public
            }
          }`,
          variables: {id: this.$store.state.project.project},
          fetchPolicy: 'network-only'
        }).then(({ data }) => {
          this.name = data.project.name
          this.isPublic = data.project.public === noneAccessType
          this.$store.commit('auth/setRedirectLogin', this.isPublic)
        }).catch(err => {
          console.error(err)
          this.$bvToast.toast(`found error: ${err.message}`, {
            variant: 'danger',
            title: 'Error'
          })
        })
    },
    updateProject(evt) {
      evt.preventDefault()
      this.$apollo.mutate({mutation: gql`
        mutation updateProject($id: String!, $name: String!){updateProject(id: $id, name: $name){id} }
        `, variables: {id: this.$store.state.project.project, name: this.name}})
        .then(({ data }) => {
          console.log('updated!')
        }).catch(err => {
          console.error(err)
          this.$bvToast.toast(`found error: ${err.message}`, {
            variant: 'danger',
            title: 'Error'
          })
        })
    }
  }
})
</script>

<style lang="scss">
.new-form-button {
  height: 6rem;
  width: 6rem;
  text-align: center;
  line-height: 50%;
  z-index: 99;
  position: fixed;
  bottom: 3rem;
  right: 3rem;
}
</style>

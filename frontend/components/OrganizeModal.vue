<template>
  <b-modal ref="organize-modal" @ok="organize" size="xl" title="Organize">
    <multiselect
      v-model="tags"
      :multiple="true"
      :options="$store.state.auth.user.tags"
      :taggable="true"
      label="name"
    />
    <multiselect
      v-model="categories"
      :multiple="true"
      :options="$store.state.auth.user.categories"
      :taggable="true"
      label="name"
    />
    <add-organize />
  </b-modal>
</template>

<script lang="js">
import Vue from 'vue'
import gql from 'graphql-tag'
import Multiselect from 'vue-multiselect'
import AddOrganize from '~/components/AddOrganize.vue'
import { clone } from '~/assets/utils'
import { validTypes } from '~/assets/config'
// edit project or form with organization
export default Vue.extend({
  name: 'Organize',
  components: {
    Multiselect,
    AddOrganize
  },
  props: {
    initialCategories: {
      type: Array,
      default: null
    },
    initialTags: {
      type: Array,
      default: null
    },
    type: {
      type: String,
      default: null,
      validator: (value) => validTypes.includes(value)
    },
    id: {
      type: String,
      default: null
    }
  },
  data() {
    return {
      tags: [],
      categories: []
    }
  },
  mounted() {
    this.tags = clone(this.initialTags)
    this.categories = clone(this.initialCategories)
  },
  methods: {
    show() {
      if (this.$refs['organize-modal']) {
        this.$refs['organize-modal'].show()
      } else {
        this.$bvToast.toast('cannot find organize modal', {
          variant: 'danger',
          title: 'Error'
        })
      }
    },
    organize(evt) {
      evt.preventDefault()
      const finished = () => {
        this.$nextTick(() => {
          if (this.$refs['organize-modal']) {
            this.$refs['organize-modal'].hide()
          }
        })
      }
      if (this.tags.length > 0 || this.categories.length > 0) {
        const tags = this.tags.map(tag => tag.name)
        const categories = this.categories.map(category => category.name)
        if (this.type === validTypes[0]) {
          // project type
          this.$apollo.mutate({mutation: gql`
            mutation updateProject($id: String!, $tags: [String!], $categories: [String!]) {
              updateProject(id: $id, tags: $tags, categories: $categories) {
                id
              }
            }
            `, variables: {
              id: this.id,
              categories,
              tags
            }})
            .then(({ data }) => {
              finished()
            }).catch(err => {
              this.$bvToast.toast(`found error: ${err.message}`, {
                variant: 'danger',
                title: 'Error'
              })
            })
        } else {
          this.$apollo.mutate({mutation: gql`
            mutation updateForm($id: String!, $tags: [String!], $categories: [String!]) {
              updateForm(id: $id, tags: $tags, categories: $categories) {
                id
              }
            }
            `, variables: {
              id: this.id,
              categories,
              tags
            }})
            .then(({ data }) => {
              finished()
            }).catch(err => {
              this.$bvToast.toast(`found error: ${err.message}`, {
                variant: 'danger',
                title: 'Error'
              })
            })
        }
      } else {
        finished()
      }
    }
  }
})
</script>

<style lang="scss"></style>

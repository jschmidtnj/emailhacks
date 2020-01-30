<template>
  <b-container v-if="!loading" fluid>
    <b-row>
      <b-col class="my-1">
        {{ `${numForms} form${numForms === 1 ? '' : 's'}` }}
        <b-form-group class="mb-0">
          <b-input-group>
            <b-form-input
              v-model="search"
              @keyup.enter.native="updatedSearchTerm"
              placeholder="Type to Search"
            />
            <b-input-group-append>
              <b-button
                :disabled="!search"
                @click="
                  search = ''
                  currentPage = 1
                  searchForms(false)
                "
              >
                Clear
              </b-button>
            </b-input-group-append>
          </b-input-group>
        </b-form-group>
      </b-col>
    </b-row>
    <b-container
      v-infinite-scroll="searchForms(true)"
      infinite-scroll-disabled="scrollDisabled"
      infinite-scroll-distance="10"
    >
      <b-card v-for="(item, index) in items" :key="`item-${index}`" no-body>
        <b-card-body>
          <b-row>
            <b-col>
              {{ item.name }}
            </b-col>
            <b-col class="text-right">
              <nuxt-link
                :to="`form/${item.id}`"
                class="btn btn-primary btn-sm no-underline"
              >
                View
              </nuxt-link>
            </b-col>
            <b-col class="text-right">
              <b-button @click="deleteForm(item)" size="sm">
                Delete
              </b-button>
            </b-col>
            <b-col class="text-right">
              <b-button
                @click="(evt) => share(evt, item.id)"
                pill
                variant="primary"
                class="share-button"
              >
                <client-only>
                  <font-awesome-icon size="md" icon="share" />
                </client-only>
              </b-button>
            </b-col>
          </b-row>
        </b-card-body>
      </b-card>
    </b-container>
    <!--share-modal ref="share-modal" :id="currentFormId" type="form" /-->
  </b-container>
</template>

<script lang="js">
import Vue from 'vue'
import gql from 'graphql-tag'
// import ShareModal from '~/components/ShareModal.vue'
const searchIntervalDuration = 750 // update search every ms after change
const defaultSort = 'updated'
export default Vue.extend({
  name: 'Forms',
  components: {
    // ShareModal
  },
  data() {
    return {
      scrollDisabled: true,
      items: [],
      currentFormId: null,
      numForms: 0,
      currentPage: 1,
      perPage: 5,
      pageOptions: [5, 10, 15],
      sortBy: null,
      sortDesc: true,
      search: '',
      loading: true,
      searchInterval: null
    }
  },
  beforeDestroy() {
    this.clearSearchInterval()
  },
  mounted() {
    if (this.$route.query) {
      if (this.$route.query.phrase) this.search = this.$route.query.phrase
      if (this.$route.query.perpage)
        this.perPage = parseInt(this.$route.query.perpage)
      if (this.$route.query.currentpage)
        this.currentPage = parseInt(this.$route.query.currentpage)
      if (this.$route.query.sortdescending)
        this.sortDesc = this.$route.query.sortdescending === 'true'
      if (this.$route.query.sortby)
        this.sortBy = this.$route.query.sortby
    }
    if (!this.sortBy) {
      this.sortBy = defaultSort
    }
    console.log(`start sort by ${this.sortBy}`)
    this.loading = false
  },
  methods: {
    share(evt, formId) {
      evt.preventDefault()
      console.log('share!')
      this.currentFormId = formId
      if (this.$refs['share-modal']) {
        this.$refs['submit-modal'].show()
      } else {
        this.$bvToast.toast('cannot find share modal', {
          variant: 'danger',
          title: 'Error'
        })
      }
    },
    deleteForm(form) {
      this.$apollo.mutate({mutation: gql`
        mutation deleteForm($id: String!){deleteForm(id: $id){id} }
        `, variables: {id: form.id}})
        .then(({ data }) => {
          this.items.splice(this.items.indexOf(form), 1)
          this.$bvToast.toast('form deleted', {
            variant: 'success',
            title: 'Success'
          })
        }).catch(err => {
          console.error(err)
          this.$bvToast.toast(`found error: ${err.message}`, {
            variant: 'danger',
            title: 'Error'
          })
        })
    },
    clearSearchInterval() {
      if (this.searchInterval) {
        clearInterval(this.searchInterval)
      }
    },
    updatedSearchTerm(evt) {
      evt.preventDefault()
      this.currentPage = 1
      this.clearSearchInterval()
      this.searchInterval = setInterval(this.searchForms(false), searchIntervalDuration)
    },
    updateCount() {
      this.$axios
        .get('/countForms', {
          params: {
            searchterm: this.search,
            tags: [].join(',tags='),
            categories: [].join(',categories=')
          }
        })
        .then(res => {
          if (res.status === 200) {
            if (res.data) {
              if (res.data.count !== null) {
                this.numForms = res.data.count
              } else {
                this.$bvToast.toast('could not find count data', {
                  variant: 'danger',
                  title: 'Error'
                })
              }
            } else {
              this.$bvToast.toast('could not get data', {
                variant: 'danger',
                title: 'Error'
              })
            }
          } else {
            this.$bvToast.toast(`status code of ${res.status}`, {
              variant: 'danger',
              title: 'Error'
            })
          }
        })
        .catch(err => {
          let message = `got error: ${err}`
          console.log(message)
          if (err.response && err.response.data) {
            message = err.response.data.message
          }
          this.$bvToast.toast(message, {
            variant: 'danger',
            title: 'Error'
          })
        })
    },
    searchForms(append) {
      this.scrollDisabled = true
      // this.updateCount()
      this.$apollo.query({
        query: gql`
          query forms($project: String!, $perpage: Int!, $page: Int!, $searchterm: String!, $sort: String!, $ascending: Boolean!, $tags: [String!]!, $categories: [String!]!) {
            forms(project: $project, perpage: $perpage, page: $page, searchterm: $searchterm, sort: $sort, ascending: $ascending, tags: $tags, categories: $categories) {
              name
              views
              id
              updated
            }
          }`,
          variables: {project: this.$store.state.project.projectId, perpage: this.perPage, page: this.currentPage - 1, searchterm: this.search, sort: this.sortBy, ascending: !this.sortDesc, tags: [], categories: []},
          fetchPolicy: 'network-only'
        })
        .then(({ data }) => {
          const forms = data.forms
          forms.forEach(form => {
            if (form.updated && form.updated.toString().length === 10) {
              form.updated = Number(form.updated) * 1000
            }
            if (form.created && form.created.toString().length === 10) {
              form.created = Number(form.created) * 1000
            }
          })
          if (append) {
            this.items.concat(forms)
            if (this.items.length > 0) {
              this.currentPage++
              this.scrollDisabled = false
            }
          } else {
            this.items = forms
            this.scrollDisabled = false
          }
          /*
            this.$nextTick(() => {
              this.$forceUpdate()
            })
          */
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
.share-button {
  height: 2rem;
  width: 2rem;
  text-align: center;
  line-height: 50%;
}
</style>

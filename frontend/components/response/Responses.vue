<template>
  <b-container fluid>
    <b-row>
      <b-col md="6" class="my-1">
        <b-form-group label-cols-sm="3" label="search" class="mb-0">
          <b-input-group>
            <b-form-input
              v-model="search"
              @keyup.enter.native="
                (evt) => {
                  evt.preventDefault()
                  currentPage = 1
                  searchResponses(false)
                }
              "
              placeholder="Type to Search"
            />
            <b-input-group-append>
              <b-button
                :disabled="!search"
                @click="
                  search = ''
                  currentPage = 1
                  searchResponses(false)
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
      v-infinite-scroll="searchResponses(true)"
      infinite-scroll-disabled="gettingData"
      infinite-scroll-distance="10"
    >
      <b-card v-for="(item, index) in items" :key="`item-${index}`" no-body>
        <b-card-body>
          <b-row>
            <b-col>
              {{ item.user }}
            </b-col>
            <b-col class="text-right">
              <nuxt-link
                :to="
                  `form/${formId ? formId : data.item.form}/response/${
                    data.item.id
                  }`
                "
                class="btn btn-primary btn-sm no-underline"
              >
                View{{
                  $store.state.auth.user.id === data.item.user ? ' + Edit' : ''
                }}
              </nuxt-link>
            </b-col>
            <b-col v-if="editAccess" class="text-right">
              <b-button @click="deleteResponse(item)" size="sm">
                Delete
              </b-button>
            </b-col>
          </b-row>
        </b-card-body>
      </b-card>
    </b-container>
  </b-container>
</template>

<script lang="js">
import Vue from 'vue'
import gql from 'graphql-tag'
const defaultSort = 'updated'
export default Vue.extend({
  name: 'Responses',
  layout: 'secure',
  props: {
    formId: {
      type: String,
      default: null
    },
    editAccess: {
      type: Boolean,
      default: false
    }
  },
  data() {
    return {
      hasEditAccess: false,
      items: [],
      gettingData: true,
      totalRows: 0,
      currentPage: 1,
      perPage: 5,
      pageOptions: [5, 10, 15],
      sortBy: null,
      sortDesc: true,
      search: ''
    }
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
    this.hasEditAccess = this.editAccess
    this.searchResponses(false)
  },
  methods: {
    deleteResponse(response) {
      this.$apollo.mutate({mutation: gql`
        mutation deleteResponse($id: String!){deleteResponse(id: $id){id} }
        `, variables: {id: response.id}})
        .then(({ data }) => {
          this.items.splice(this.items.indexOf(response), 1)
          this.$bvToast.toast('response deleted', {
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
    updateCount() {
      const getParams = {
        searchterm: this.search,
        tags: [].join(',tags='),
        categories: [].join(',categories=')
      }
      if (this.formId) {
        getParams.form = this.formId
      }
      this.$axios
        .get('/countResponses', {
          params: getParams
        })
        .then(res => {
          if (res.status === 200) {
            if (res.data) {
              if (res.data.count !== null) {
                this.totalRows = res.data.count
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
          if (err.response && err.response.data) {
            message = err.response.data.message
          }
          this.$bvToast.toast(message, {
            variant: 'danger',
            title: 'Error'
          })
        })
    },
    searchResponses(append) {
      this.gettingData = true
      this.updateCount()
      const formIdVal = this.formId ? this.formId : null
      this.$apollo.query({
        query: gql`
          query responses($form: String, $perpage: Int!, $page: Int!, $searchterm: String!, $sort: String!, $ascending: Boolean!) {
            responses(form: $form, perpage: $perpage, page: $page, searchterm: $searchterm, sort: $sort, ascending: $ascending) {
              ${this.formId ? 'user' : 'project, form'},
              views,
              id,
              updated
            }
          }`,
          variables: {form: formIdVal, perpage: this.perPage, page: this.currentPage - 1, searchterm: this.search, sort: this.sortBy, ascending: !this.sortDesc, tags: [], categories: []},
          fetchPolicy: 'network-only'
        }).then(({ data }) => {
          const responses = data.responses
          console.log(responses)
          responses.forEach(response => {
            if (response.updated && response.updated.toString().length === 10) {
              response.updated = Number(response.updated) * 1000
            }
            if (response.created && response.created.toString().length === 10) {
              response.created = Number(response.created) * 1000
            }
          })
          if (append) {
            this.items.concat(responses)
          } else {
            this.items = responses
          }
          this.$nextTick(() => {
            this.$forceUpdate()
          })
          this.gettingData = false
        }).catch(err => {
          console.error(err)
          this.$bvToast.toast(`found error: ${err.message}`, {
            variant: 'danger',
            title: 'Error'
          })
          this.gettingData = false
        })
    }
  }
})
</script>

<style lang="scss"></style>

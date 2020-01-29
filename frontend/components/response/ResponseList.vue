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
                  searchResponses()
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
                  searchResponses()
                "
              >
                Clear
              </b-button>
            </b-input-group-append>
          </b-input-group>
        </b-form-group>
      </b-col>
      <b-col md="6" class="my-1">
        <b-form-group label-cols-sm="3" label="Sort" class="mb-0">
          <b-input-group>
            <b-form-select
              v-model="sortBy"
              :options="sortOptions"
              @change="
                currentPage = 1
                searchResponses()
              "
            >
              <option slot="first" :value="null">
                -- none --
              </option>
            </b-form-select>
            <b-form-select
              slot="prepend"
              v-model="sortDesc"
              :disabled="!sortBy"
              @change="
                currentPage = 1
                searchResponses()
              "
            >
              <option :value="false">
                Asc
              </option>
              <option :value="true">
                Desc
              </option>
            </b-form-select>
          </b-input-group>
        </b-form-group>
      </b-col>
      <b-col md="6" class="my-1">
        <b-form-group label-cols-sm="3" label="Per page" class="mb-0">
          <b-form-select
            v-model="perPage"
            :options="pageOptions"
            @change="
              currentPage = 1
              searchResponses()
            "
          />
        </b-form-group>
      </b-col>
    </b-row>
    <b-table
      :items="items"
      :fields="fields"
      :no-local-sorting="true"
      @sort-changed="sort"
      show-empty
      stacked="md"
    >
      <template v-if="formId" v-slot:cell(user)="data">
        {{ data.value }}
      </template>
      <template v-else v-slot:cell(form)="data">
        {{ data.value }}
      </template>
      <template v-slot:cell(updated)="data">
        {{ formatDate(data.value) }}
      </template>
      <template v-slot:cell(views)="data">
        {{ data.value }}
      </template>
      <template v-slot:cell(actions)="data">
        <nuxt-link
          :to="
            `form/${formId ? formId : data.item.form}/response/${data.item.id}`
          "
          class="btn btn-primary btn-sm no-underline"
        >
          View{{
            $store.state.auth.user.id === data.item.user ? ' + Edit' : ''
          }}
        </nuxt-link>
        <b-button
          v-if="editAccess"
          @click="deleteResponse(data.item)"
          size="sm"
        >
          Delete
        </b-button>
      </template>
    </b-table>
    <b-row>
      <b-col md="6" class="my-1">
        <b-pagination
          v-model="currentPage"
          :total-rows="totalRows"
          :per-page="perPage"
          @change="
            (newpage) => {
              currentPage = newpage
              searchResponses()
            }
          "
          class="my-0"
        />
      </b-col>
    </b-row>
    <nuxt-link
      v-if="formId"
      :to="`form/${formId}/view`"
      class="btn btn-primary btn-sm no-underline mt-4"
    >
      Create New Response
    </nuxt-link>
  </b-container>
</template>

<script lang="js">
import Vue from 'vue'
import { formatRelative } from 'date-fns'
import gql from 'graphql-tag'
// Advanced search for responses
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
      items: [],
      fields: [
        {
          key: this.formId ? 'user' : 'form',
          label: this.formId ? 'User Id' : 'Form Id',
          sortable: true
        },
        {
          key: 'updated',
          label: 'Updated',
          sortable: true,
          class: 'text-center'
        },
        {
          key: 'views',
          label: 'Views',
          sortable: true,
          sortDirection: 'desc'
        },
        {
          key: 'actions',
          label: 'Actions',
          sortable: false
        }
      ],
      totalRows: 0,
      currentPage: 1,
      perPage: 5,
      pageOptions: [5, 10, 15],
      sortBy: null,
      sortDesc: true,
      search: ''
    }
  },
  computed: {
    sortOptions() {
      // Create an options list from our fields
      return this.fields
        .filter(f => f.sortable)
        .map(f => {
          return { text: f.label, value: f.key }
        })
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
      if (
        this.$route.query.sortby &&
        this.fields.some(field => field.key === this.$route.query.sortby)
      )
        this.sortBy = this.$route.query.sortby
    }
    this.searchResponses()
  },
  methods: {
    sort(ctx) {
      this.sortBy = ctx.sortBy //   ==> Field key for sorting by (or null for no sorting)
      this.sortDesc = ctx.sortDesc // ==> true if sorting descending, false otherwise
      this.currentPage = 1
      this.searchResponses()
    },
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
    searchResponses() {
      this.updateCount()
      const sort = this.sortBy ? this.sortBy : this.sortOptions[0].value
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
          variables: {form: formIdVal, perpage: this.perPage, page: this.currentPage - 1, searchterm: this.search, sort, ascending: !this.sortDesc, tags: [], categories: []},
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
          this.items = responses
          this.$nextTick(() => {
            this.$forceUpdate()
          })
        }).catch(err => {
          console.error(err)
          this.$bvToast.toast(`found error: ${err.message}`, {
            variant: 'danger',
            title: 'Error'
          })
        })
    },
    formatDate(dateUTC) {
      return formatRelative(dateUTC, new Date())
    }
  }
})
</script>

<style lang="scss"></style>

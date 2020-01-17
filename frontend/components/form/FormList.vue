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
                  searchForms()
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
                  searchForms()
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
                searchForms()
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
                searchForms()
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
              searchForms()
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
      <template v-slot:cell(name)="data">
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
            `/project/${projectId ? projectId : data.item.project}/form/${
              data.item.id
            }/view`
          "
          class="btn btn-primary btn-sm no-underline"
        >
          View
        </nuxt-link>
        <nuxt-link
          :to="
            `/project/${projectId ? projectId : data.item.project}/form/${
              data.item.id
            }/edit`
          "
          class="btn btn-primary btn-sm no-underline"
        >
          Edit
        </nuxt-link>
        <b-button @click="deleteForm(data.item)" size="sm">
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
              searchForms()
            }
          "
          class="my-0"
        />
      </b-col>
    </b-row>
  </b-container>
</template>

<script lang="js">
import Vue from 'vue'
import { formatRelative } from 'date-fns'
import gql from 'graphql-tag'
export default Vue.extend({
  name: 'Forms',
  props: {
    projectId: {
      type: String,
      default: null
    }
  },
  data() {
    return {
      items: [],
      fields: [
        {
          key: 'name',
          label: 'Name',
          sortable: false
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
    this.searchForms()
  },
  methods: {
    deleteForm(form) {
      this.$apollo.mutate({mutation: gql`
        mutation deleteForm($id: String!){deleteForm(id: $id){id} }
        `, variables: {id: form.id}})
        .then(({ data }) => {
          this.items.splice(this.items.indexOf(form), 1)
          this.$toasted.global.success({
            message: 'form deleted'
          })
        }).catch(err => {
          console.error(err)
          this.$toasted.global.error({
            message: `found error: ${err.message}`
          })
        })
    },
    sort(ctx) {
      this.sortBy = ctx.sortBy //   ==> Field key for sorting by (or null for no sorting)
      this.sortDesc = ctx.sortDesc // ==> true if sorting descending, false otherwise
      this.currentPage = 1
      this.searchForms()
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
                this.totalRows = res.data.count
                console.log(res.data.count)
              } else {
                this.$toasted.global.error({
                  message: 'could not find count data'
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
          let message = `got error: ${err}`
          if (err.response && err.response.data) {
            message = err.response.data.message
          }
          this.$toasted.global.error({
            message
          })
        })
    },
    searchForms() {
      this.updateCount()
      const sort = this.sortBy ? this.sortBy : this.sortOptions[0].value
      console.log(`sort by ${sort}`)
      this.$apollo.query({query: gql`
        query forms($perpage: Int!, $page: Int!, $searchterm: String!, $sort: String!, $ascending: Boolean!, $tags: [String!]!, $categories: [String!]!)
          {forms(perpage: $perpage, page: $page, searchterm: $searchterm, sort: $sort, ascending: $ascending, tags: $tags, categories: $categories){
            name
            views
            id
            updated
            project
           }
          }
        `, variables: {perpage: this.perPage, page: this.currentPage - 1, searchterm: this.search, sort, ascending: !this.sortDesc, tags: [], categories: []}})
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
          this.items = forms
          this.$nextTick(() => {
            this.$forceUpdate()
          })
        }).catch(err => {
          console.error(err)
          this.$toasted.global.error({
            message: `found error: ${err.message}`
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

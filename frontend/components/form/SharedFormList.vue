<template>
  <b-container class="mt-4">
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
                    searchProjects()
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
                    searchProjects()
                  "
                >
                  Clear
                </b-button>
              </b-input-group-append>
            </b-input-group>
          </b-form-group>
        </b-col>
      </b-row>
      <!-- replace table with listed data: -->
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
            :to="`/project/${data.item.id}`"
            class="btn btn-primary btn-sm no-underline"
          >
            View
          </nuxt-link>
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
                searchProjects()
              }
            "
            class="my-0"
          />
        </b-col>
      </b-row>
    </b-container>
    <nuxt-link to="/project" class="btn btn-primary btn-sm no-underline mt-4">
      Create New Project
    </nuxt-link>
  </b-container>
</template>

<script lang="js">
import Vue from 'vue'
import { formatRelative } from 'date-fns'
import gql from 'graphql-tag'
// TODO - make tiles for the form dropdowns (search when drop down is pressed)
export default Vue.extend({
  name: 'ProjectList',
  layout: 'secure',
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
    this.searchProjects()
  },
  methods: {
    sort(ctx) {
      this.sortBy = ctx.sortBy //   ==> Field key for sorting by (or null for no sorting)
      this.sortDesc = ctx.sortDesc // ==> true if sorting descending, false otherwise
      this.currentPage = 1
      this.searchProjects()
    },
    updateCount() {
      this.$axios
        .get('/countProjects', {
          params: {
            searchterm: this.search,
            tags: [].join(',tags='),
            categories: [].join(',categories='),
            onlyshared: true
          }
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
    searchProjects() {
      this.updateCount()
      const sort = this.sortBy ? this.sortBy : this.sortOptions[0].value
      this.$apollo.query({
        query: gql`
          query projects($perpage: Int!, $page: Int!, $searchterm: String!, $sort: String!, $ascending: Boolean!, $tags: [String!]!, $categories: [String!]!) {
            projects(onlyshared: true, perpage: $perpage, page: $page, searchterm: $searchterm, sort: $sort, ascending: $ascending, tags: $tags, categories: $categories) {
              name,
              views,
              id,
              updated
            }
          }`,
          variables: {perpage: this.perPage, page: this.currentPage - 1, searchterm: this.search, sort, ascending: !this.sortDesc, tags: [], categories: []},
          fetchPolicy: 'network-only'
        }).then(({ data }) => {
          const projects = data.projects
          projects.forEach(project => {
            if (project.updated && project.updated.toString().length === 10) {
              project.updated = Number(project.updated) * 1000
            }
            if (project.created && project.created.toString().length === 10) {
              project.created = Number(project.created) * 1000
            }
          })
          this.items = projects
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

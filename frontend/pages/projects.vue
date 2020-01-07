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
        <b-col md="6" class="my-1">
          <b-form-group label-cols-sm="3" label="Sort" class="mb-0">
            <b-input-group>
              <b-form-select
                v-model="sortBy"
                :options="sortOptions"
                @change="
                  currentPage = 1
                  searchProjects()
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
                  searchProjects()
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
                searchProjects()
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
            :to="`/project/${data.item.id}`"
            class="btn btn-primary btn-sm no-underline"
          >
            View
          </nuxt-link>
          <b-button @click="deleteProject(data.item)" size="sm">
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
// @ts-ignore
const seo = JSON.parse(process.env.seoconfig)
export default Vue.extend({
  name: 'Projects',
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
  // @ts-ignore
  head() {
    const title = 'Search Projects'
    const description = 'search for projects, by name, views, etc'
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
    deleteProject(project) {
      this.$apollo.mutate({mutation: gql`
        mutation deleteProject($id: String!){deleteProject(id: $id){id} }
        `, variables: {id: project.id}})
        .then(({ data }) => {
          this.items.splice(this.items.indexOf(project), 1)
          this.$toasted.global.success({
            message: 'project deleted'
          })
        }).catch(err => {
          console.error(err)
          this.$toasted.global.error({
            message: `found error: ${err.message}`
          })
        })
    },
    updateCount() {
      this.$axios
        .get('/countProjects', {
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
    searchProjects() {
      this.updateCount()
      const sort = this.sortBy ? this.sortBy : this.sortOptions[0].value
      this.$apollo.query({query: gql`
        query projects($perpage: Int!, $page: Int!, $searchterm: String!, $sort: String!, $ascending: Boolean!, $tags: [String!]!, $categories: [String!]!)
          {projects(perpage: $perpage, page: $page, searchterm: $searchterm, sort: $sort, ascending: $ascending, tags: $tags, categories: $categories){name, views, id, updated} }
        `, variables: {perpage: this.perPage, page: this.currentPage - 1, searchterm: this.search, sort, ascending: !this.sortDesc, tags: [], categories: []}})
        .then(({ data }) => {
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

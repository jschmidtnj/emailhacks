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
          :to="`form/${data.item.id}/view`"
          class="btn btn-primary btn-sm no-underline"
        >
          View
        </nuxt-link>
        <nuxt-link
          :to="`form/${data.item.id}/responses`"
          class="btn btn-primary btn-sm no-underline"
        >
          Responses
        </nuxt-link>
        <nuxt-link
          :to="`form/${data.item.id}/edit`"
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
const seo = JSON.parse(process.env.seoconfig)
// Advanced search in forms
export default Vue.extend({
  name: 'Forms',
  // @ts-ignore
  layout: 'secure',
  // @ts-ignore
  head() {
    const title = 'Dashboard'
    const description = 'project dashboard'
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
    searchForms() {
      this.updateCount()
      const sort = this.sortBy ? this.sortBy : this.sortOptions[0].value
      console.log(`sort by ${sort}`)
      this.$apollo.query({
        query: gql`
          query forms(project: String!, $perpage: Int!, $page: Int!, $searchterm: String!, $sort: String!, $ascending: Boolean!, $tags: [String!]!, $categories: [String!]!) {
            forms(project: $project, perpage: $perpage, page: $page, searchterm: $searchterm, sort: $sort, ascending: $ascending, tags: $tags, categories: $categories) {
              name
              views
              id
              updated
              project
            }
          }`,
          variables: {project: this.$store.state.project.projectId, perpage: this.perPage, page: this.currentPage - 1, searchterm: this.search, sort, ascending: !this.sortDesc, tags: [], categories: []},
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
          this.items = forms
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

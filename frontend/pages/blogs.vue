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
                  searchBlogs()
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
                  searchBlogs()
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
                searchBlogs()
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
                searchBlogs()
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
              searchBlogs()
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
      <template v-slot:cell(title)="data">
        {{ data.value }}
      </template>
      <template v-slot:cell(author)="data">
        {{ data.value }}
      </template>
      <template v-slot:cell(updated)="data">
        {{ formatDate(data.value) }}
      </template>
      <template v-slot:cell(views)="data">
        {{ data.value }}
      </template>
      <template v-slot:cell(read)="data">
        <nuxt-link
          :to="`/blog/${data.item.id}`"
          class="btn btn-primary btn-sm no-underline"
        >
          Read
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
              searchBlogs()
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
import gql from 'graphql-tag'
import { formatRelative } from 'date-fns'
import { adminTypes } from '~/assets/config'
// @ts-ignore
const seo = JSON.parse(process.env.seoconfig)
export default Vue.extend({
  name: 'Blogs',
  data() {
    return {
      type: 'blog',
      items: [],
      fields: [
        {
          key: 'title',
          label: 'Title',
          sortable: true,
          sortDirection: 'desc'
        },
        {
          key: 'author',
          label: 'Author',
          sortable: true
        },
        {
          key: 'date',
          label: 'Date',
          sortable: true,
          class: 'text-center'
        },
        {
          key: 'views',
          label: 'Views',
          sortable: true,
          class: 'text-center'
        },
        {
          key: 'read',
          label: 'Read',
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
    const title = 'Search Blogs'
    const description = 'search for blogs, by name, views, etc'
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
    this.searchBlogs()
  },
  methods: {
    sort(ctx) {
      this.sortBy = ctx.sortBy //   ==> Field key for sorting by (or null for no sorting)
      this.sortDesc = ctx.sortDesc // ==> true if sorting descending, false otherwise
      this.currentPage = 1
      this.searchBlogs()
    },
    updateCount() {
      this.$axios
        .get('/countBlogs', {
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
    searchBlogs() {
      this.updateCount()
      const sort = this.sortBy ? this.sortBy : this.sortOptions[0].value
      const useCache = this.$store.state.auth.user && adminTypes.includes(this.$store.state.auth.user.type)
      this.$apollo.query({
        query: gql`
          query blogs($perpage: Int!, $page: Int!, $searchterm: String!, $sort: String!, $ascending: Boolean!, $tags: [String!]!, $categories: [String!]!, $cache: Boolean!) {
            blogs(perpage: $perpage, page: $page, searchterm: $searchterm, sort: $sort, ascending: $ascending, tags: $tags, categories: $categories, cache: $cache) {
              title
              views
              id
              author
              updated
            }
          }`,
          variables: {
              perpage: this.perPage,
              page: this.currentPage - 1,
              searchterm: this.search,
              sort,
              ascending: !this.sortDesc,
              tags: [],
              categories: [],
              cache: useCache
            },
            fetchPolicy: useCache ? 'cache-first' : 'network-only'
          }).then(({ data }) => {
            const blogs = data.blogs
            blogs.forEach(blog => {
              if (blog.updated && blog.updated.toString().length === 10) {
                blog.updated = Number(blog.updated) * 1000
              }
              if (blog.created && blog.created.toString().length === 10) {
                blog.created = Number(blog.created) * 1000
              }
            })
            this.items = blogs
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

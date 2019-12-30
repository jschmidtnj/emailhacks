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
            ></b-form-input>
            <b-input-group-append>
              <b-button
                :disabled="!search"
                @click="
                  search = ''
                  currentPage = 1
                  searchForms()
                "
                >Clear</b-button
              >
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
              <option slot="first" :value="null">-- none --</option>
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
              <option :value="false">Asc</option>
              <option :value="true">Desc</option>
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
          ></b-form-select>
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
      <template v-slot:cell(name)="data">{{ data.value }}</template>
      <template v-slot:cell(updated)="data">{{
        formatDate(data.value)
      }}</template>
      <template v-slot:cell(views)="data">{{ data.value }}</template>
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
        <b-button @click="deleteForm(data.item)" size="sm">Delete</b-button>
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
        ></b-pagination>
      </b-col>
    </b-row>
  </b-container>
</template>

<script lang="js">
import Vue from 'vue'
import { formatRelative } from 'date-fns'
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
      sortDesc: false,
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
    this.searchForms(this.currentPage)
  },
  methods: {
    deleteForm(form) {
      this.$axios
        .delete('/graphql', {
          params: {
            query: `mutation{deleteForm(id:"${encodeURIComponent(form.id)}"){id}}`
          }
        })
        .then(res => {
          if (res.status === 200) {
            if (res.data) {
              if (res.data.data && res.data.data.deleteForm) {
                this.items.splice(this.items.indexOf(form), 1)
                this.$toasted.global.success({
                  message: 'form deleted'
                })
              } else if (res.data.errors) {
                this.$toasted.global.error({
                  message: `found errors: ${JSON.stringify(res.data.errors)}`
                })
              } else {
                this.$toasted.global.error({
                  message: 'could not find data or errors'
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
          this.$toasted.global.error({
            message: err
          })
        })
    },
    sort(ctx) {
      this.sortBy = ctx.sortBy //   ==> Field key for sorting by (or null for no sorting)
      this.sortDesc = ctx.sortDesc // ==> true if sorting descending, false otherwise
      this.currentPage = 1
      this.searchForms(this.currentPage)
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
      this.$axios
        .get('/graphql', {
          params: {
            query: `{forms(${
              this.projectId ? `project:"${encodeURIComponent(this.projectId)}",` : ''
            }perpage:${
              this.perPage
            },page:${this.currentPage - 1},searchterm:"${encodeURIComponent(
              this.search
            )}",sort:"${encodeURIComponent(sort)}",ascending:${!this
              .sortDesc},tags:${JSON.stringify([])},categories:${JSON.stringify(
              []
            )}){name views${ this.projectId ? '' : ' project' } id updated}}`
          }
        })
        .then(res => {
          if (res.status === 200) {
            if (res.data) {
              if (res.data.data && res.data.data.forms) {
                const forms = res.data.data.forms
                forms.forEach(form => {
                  Object.keys(form).forEach(key => {
                    if (typeof form[key] === 'string')
                      form[key] = decodeURIComponent(form[key])
                  })
                  if (form.updated) {
                    form.updated = Number(form.updated) * 1000
                  }
                  if (form.created) {
                    form.created = Number(form.created) * 1000
                  }
                })
                this.items = forms
              } else if (res.data.errors) {
                this.$toasted.global.error({
                  message: `found errors: ${JSON.stringify(res.data.errors)}`
                })
              } else {
                this.$toasted.global.error({
                  message: 'could not find data or errors'
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
    formatDate(dateUTC) {
      return formatRelative(dateUTC, new Date())
    }
  }
})
</script>

<style lang="scss"></style>

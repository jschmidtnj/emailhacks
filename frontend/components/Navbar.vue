<template>
  <b-navbar toggleable="lg" type="dark" variant="info">
    <nuxt-link v-if="!loggedIn" to="/" class="no-underline">
      <b-navbar-brand href="/">
        Mail Pear
      </b-navbar-brand>
    </nuxt-link>
    <b-navbar-toggle target="nav-collapse" />
    <b-collapse id="nav-collapse" is-nav>
      <b-navbar-nav>
        <nuxt-link v-if="!loggedIn" to="/about" class="no-underline">
          <b-nav-item href="/about">
            About
          </b-nav-item>
        </nuxt-link>
        <nuxt-link v-if="!loggedIn" to="/blogs" class="no-underline">
          <b-nav-item href="/blogs">
            Blogs
          </b-nav-item>
        </nuxt-link>
        <nuxt-link v-if="!loggedIn" to="/signup" class="no-underline">
          <b-nav-item href="/signup">
            Signup
          </b-nav-item>
        </nuxt-link>
        <nuxt-link v-if="!loggedIn" to="/login" class="no-underline">
          <b-nav-item href="/login">
            Login
          </b-nav-item>
        </nuxt-link>
        <nuxt-link v-if="loggedIn" to="/dashboard" class="no-underline">
          <b-nav-item href="/dashboard">
            Dashboard
          </b-nav-item>
        </nuxt-link>
        <multiselect
          v-if="loggedIn"
          v-model="project"
          :options="projectOptions"
          :multiple="false"
          :taggable="false"
          :searchable="true"
          :loading="loadingProjects"
          :internal-search="false"
          :clear-on-select="false"
          @search-change="updatedProjectSearchTerm"
          @select="changeProject"
          label="name"
        />
        <b-nav-item
          v-if="loggedIn && upgrade"
          @click="showUpgrade"
          class="no-underline"
          href="#"
        >
          Upgrade
        </b-nav-item>
        <nuxt-link
          v-if="loggedIn && hasCartItems"
          to="/checkout"
          class="no-underline"
        >
          <b-nav-item href="/checkout">
            <client-only>
              <font-awesome-icon size="sm" icon="share" />
            </client-only>
            Cart
          </b-nav-item>
        </nuxt-link>
      </b-navbar-nav>
      <b-navbar-nav v-if="loggedIn" class="ml-auto">
        <b-nav-item-dropdown right>
          <template slot="button-content">
            <em>User</em>
          </template>
          <nuxt-link to="/profile" class="no-underline">
            <b-dropdown-item href="/profile">
              Profile
            </b-dropdown-item>
          </nuxt-link>
          <b-dropdown-item @click="logout" href="#">
            Sign Out
          </b-dropdown-item>
        </b-nav-item-dropdown>
      </b-navbar-nav>
    </b-collapse>
    <plans-modal ref="plans-modal" v-if="loggedIn" />
  </b-navbar>
</template>

<script lang="js">
import Vue from 'vue'
import gql from 'graphql-tag'
import Multiselect from 'vue-multiselect'
import { plans } from '~/assets/config'
import PlansModal from '~/components/PlansModal.vue'
const defaultPerPage = 10
const defaultSortBy = 'updated'
const searchIntervalDuration = 750 // update search every ms after change
export default Vue.extend({
  name: 'Navbar',
  components: {
    PlansModal,
    Multiselect
  },
  data() {
    return {
      project: {
        id: this.$store.state.project.projectId,
        name: this.$store.state.project.projectName
      },
      projectOptions: [],
      loadingProjects: false,
      firstTime: true,
      searchInterval: null
    }
  },
  computed: {
    loggedIn() {
      return this.$store.state.auth && this.$store.state.auth.loggedIn
    },
    upgrade() {
      return this.$store.state.auth.user && (!this.$store.state.auth.user.plan || this.$store.state.auth.user.plan === plans[0])
    },
    hasCartItems() {
      return this.$store.state.purchase.plan || this.$store.state.purchase.products.length > 0
    }
  },
  beforeDestroy() {
    this.clearSearchInterval()
  },
  methods: {
    clearSearchInterval() {
      if (this.searchInterval) {
        clearInterval(this.searchInterval)
      }
    },
    updatedProjectSearchTerm(searchterm) {
      this.loadingProjects = true
      this.clearSearchInterval()
      this.searchInterval = setInterval(this.searchProject(searchterm), searchIntervalDuration)
    },
    searchProject(searchterm) {
      this.$apollo.query({
          query: gql`
            query projects($perpage: Int!, $page: Int!, $searchterm: String!, $sort: String!, $ascending: Boolean!, $tags: [String!]!, $categories: [String!]!) {
              projects(perpage: $perpage, page: $page, searchterm: $searchterm, sort: $sort, ascending: $ascending, tags: $tags, categories: $categories) {
                id
                name
              }
            }
          `,
          variables: {perpage: defaultPerPage, page: 0, searchterm, sort: defaultSortBy, ascending: false, tags: [], categories: []},
          fetchPolicy: 'network-only'
        })
        .then(({ data }) => {
          this.projectOptions = data.projects
          this.loadingProjects = false
        })
        .catch((err) => {
          console.error(err)
          this.$bvToast.toast(err.message, {
            variant: 'danger',
            title: 'Error'
          })
          this.projectOptions = []
          this.loadingProjects = false
        })
    },
    changeProject(selectedProject) {
      this.$store.commit('project/setProjectId', selectedProject.id)
      this.$store.commit('project/setProjectName', selectedProject.name)
    },
    showUpgrade(evt) {
      evt.preventDefault()
      console.log('upgrade!')
      if (this.$refs['plans-modal']) {
        this.$refs['plans-modal'].show()
      } else {
        this.$bvToast.toast('cannot find plans modal', {
          variant: 'danger',
          title: 'Error'
        })
      }
    },
    logout(evt) {
      evt.preventDefault()
      this.$store.dispatch('auth/logout').then(() => {
        if (
          this.$nuxt.$data.layoutName === 'secure' ||
          this.$nuxt.$data.layoutName === 'admin' ||
          this.$store.state.auth && this.$store.state.auth.redirectLogin
        ) {
          if (this.$store.state.auth.redirectLogin) {
            this.$store.commit('auth/setRedirectLogin', false)
          }
          this.$router.push({
            path: '/login'
          })
        }
      }).catch(err => {
        this.$bvToast.toast(err, {
          variant: 'danger',
          title: 'Error'
        })
      })

    }
  }
})
</script>

<style src="vue-multiselect/dist/vue-multiselect.min.css"></style>

<style lang="scss"></style>

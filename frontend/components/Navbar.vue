<template>
  <b-navbar toggleable="lg" type="dark" variant="info">
    <nuxt-link to="/" class="no-underline">
      <b-navbar-brand href="/">Mail Pear</b-navbar-brand>
    </nuxt-link>
    <b-navbar-toggle target="nav-collapse"></b-navbar-toggle>
    <b-collapse id="nav-collapse" is-nav>
      <b-navbar-nav>
        <nuxt-link v-if="!loggedIn" to="/about" class="no-underline">
          <b-nav-item href="/about">About</b-nav-item>
        </nuxt-link>
        <nuxt-link v-if="!loggedIn" to="/blogs" class="no-underline">
          <b-nav-item href="/blogs">Blogs</b-nav-item>
        </nuxt-link>
        <nuxt-link v-if="!loggedIn" to="/signup" class="no-underline">
          <b-nav-item href="/signup">Signup</b-nav-item>
        </nuxt-link>
        <nuxt-link v-if="!loggedIn" to="/login" class="no-underline">
          <b-nav-item href="/login">Login</b-nav-item>
        </nuxt-link>
        <nuxt-link v-if="loggedIn" to="/dashboard" class="no-underline">
          <b-nav-item href="/dashboard">Dashboard</b-nav-item>
        </nuxt-link>
        <nuxt-link v-if="loggedIn" to="/form" class="no-underline">
          <b-nav-item href="/form">Form</b-nav-item>
        </nuxt-link>
      </b-navbar-nav>
      <b-navbar-nav v-if="loggedIn" class="ml-auto">
        <b-nav-item-dropdown right>
          <template slot="button-content">
            <em>User</em>
          </template>
          <nuxt-link to="/profile" class="no-underline">
            <b-dropdown-item href="/profile">Profile</b-dropdown-item>
          </nuxt-link>
          <b-dropdown-item @click="logout" href="#">
            Sign Out
          </b-dropdown-item>
        </b-nav-item-dropdown>
      </b-navbar-nav>
    </b-collapse>
  </b-navbar>
</template>

<script lang="js">
import Vue from 'vue'
export default Vue.extend({
  name: 'Navbar',
  data() {
    return {}
  },
  computed: {
    loggedIn() {
      return this.$store.state.auth && this.$store.state.auth.loggedIn
    }
  },
  methods: {
    logout(evt) {
      evt.preventDefault()
      this.$store.commit('auth/logout')
      if (
        this.$nuxt.$data.layoutName === 'secure' ||
        this.$nuxt.$data.layoutName === 'admin'
      ) {
        this.$router.push({
          path: '/login'
        })
      }
    }
  }
})
</script>

<style lang="scss"></style>

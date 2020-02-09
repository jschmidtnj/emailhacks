<template>
  <b-card title="Profile">
    <p>user: {{ this.$store.state.auth.user }}</p>
    <p>token: {{ this.$store.state.auth.token }}</p>
    <b-btn @click="logout">
      Logout
    </b-btn>
    <b-btn @click="deleteAccount">
      Delete
    </b-btn>
    <p>current plan: {{ currentPlan }}</p>
    <b-btn @click="showChangePlan">
      Change Plan
    </b-btn>
    <!--organize-edit /-->
    <plans-modal ref="plans-modal-profile" />
  </b-card>
</template>

<script lang="js">
import Vue from 'vue'
import gql from 'graphql-tag'
import PlansModal from '~/components/PlansModal.vue'
// import OrganizeEdit from '~/components/profile/OrganizeEdit.vue'
// @ts-ignore
const seo = JSON.parse(process.env.seoconfig)
export default Vue.extend({
  // @ts-ignore
  layout: 'secure',
  components: {
    // OrganizeEdit,
    PlansModal
  },
  computed: {
    currentPlan() {
      if (!this.$store.state.auth.user.plan) {
        return null
      }
      const product = this.$store.state.purchase.productOptions.find(product => product.id === this.$store.state.auth.user.plan)
      if (!product) {
        return null
      }
      return product.name
    }
  },
  // @ts-ignore
  head() {
    const title = 'Account'
    const description = `your account: ${
      this.$store.state.auth.user
        ? this.$store.state.auth.user.email
        : 'logging out'
    }`
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
  methods: {
    showChangePlan(evt) {
      evt.preventDefault()
      console.log('change plan.')
      if (this.$refs['plans-modal-profile']) {
        this.$refs['plans-modal-profile'].show()
      } else {
        this.$bvToast.toast('cannot find plans modal', {
          variant: 'danger',
          title: 'Error'
        })
      }
    },
    logout(evt) {
      if (evt) {
        evt.preventDefault()
      }
      this.$store.dispatch('auth/logout').then(() => {
        this.$router.push({
          path: '/login'
        })
      }).catch(err => {
        this.$bvToast.toast(err, {
          variant: 'danger',
          title: 'Error'
        })
      })
    },
    deleteAccount(evt) {
      evt.preventDefault()
      this.$apollo.mutate({mutation: gql`
        mutation deleteAccount {
          deleteAccount {
            id
          }
        }
        `, variables: {}})
        .then(({ data }) => {
          this.$bvToast.toast('account deleted', {
            variant: 'success',
            title: 'Success'
          })
          this.logout()
        }).catch(err => {
          console.error(err)
          this.$bvToast.toast(`found error: ${err.message}`, {
            variant: 'danger',
            title: 'Error'
          })
        })
    }
  }
})
</script>

<style lang="scss"></style>

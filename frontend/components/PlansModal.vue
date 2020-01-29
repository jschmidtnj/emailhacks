<template>
  <b-modal ref="plans-modal" size="xl" title="Share">
    <b-container v-if="!loading">
      <h2>{{ upgrade ? 'Upgrade' : 'Change' }} your account</h2>
      <b-form-checkbox
        v-model="annual"
        class="pull-right"
        name="allow-multiple"
        switch
      >
        {{ annual ? 'Annual' : 'Monthly' }}
      </b-form-checkbox>
      <b-row>
        <div :setProduct="(product = getProduct(plans[0]))">
          <div v-if="product" :setPlan="(plan = getPlan(product))">
            <b-col v-if="product && plan">
              <b-card no-body>
                <b-card-body>
                  <h3>free plan</h3>
                  <b-list-group>
                    <b-list-group-item
                      >max storage: {{ product.maxstorage }}</b-list-group-item
                    >
                    <b-list-group-item
                      >max projects:
                      {{ product.maxprojects }}</b-list-group-item
                    >
                    <b-list-group-item
                      >max projects:
                      {{ product.maxprojects }}</b-list-group-item
                    >
                  </b-list-group>
                  <h3>price: {{ getPrice(plan.amount) }}</h3>
                  <b-btn @click="cancelSubscription" :disabled="upgrade">
                    Select
                  </b-btn>
                </b-card-body>
              </b-card>
            </b-col>
          </div>
        </div>
        <div :setProduct="(product = getProduct(plans[1]))">
          <div v-if="product" :setPlan="(plan = getPlan(product))">
            <b-col v-if="plan">
              <b-card no-body>
                <b-card-body>
                  <h3>business plan</h3>
                  <b-list-group>
                    <b-list-group-item
                      >max storage: {{ product.maxstorage }}</b-list-group-item
                    >
                    <b-list-group-item
                      >max projects:
                      {{ product.maxprojects }}</b-list-group-item
                    >
                    <b-list-group-item
                      >max projects:
                      {{ product.maxprojects }}</b-list-group-item
                    >
                  </b-list-group>
                  <h3>price: {{ getPrice(plan.amount) }}</h3>
                  <b-btn
                    @click="
                      (evt) =>
                        purchase(evt, {
                          product: product.id,
                          interval: plan.interval
                        })
                    "
                    :disabled="$store.state.user.plan === plans[1]"
                  >
                    Select
                  </b-btn>
                </b-card-body>
              </b-card>
            </b-col>
          </div>
        </div>
        <div :setProduct="(product = getProduct(plans[2]))">
          <div v-if="product" :setPlan="(plan = getPlan(product))">
            <b-col v-if="plan">
              <b-card no-body>
                <b-card-body>
                  <h3>enterprise plan</h3>
                  <b-list-group>
                    <b-list-group-item
                      >max storage: {{ product.maxstorage }}</b-list-group-item
                    >
                    <b-list-group-item
                      >max projects:
                      {{ product.maxprojects }}</b-list-group-item
                    >
                    <b-list-group-item
                      >max projects:
                      {{ product.maxprojects }}</b-list-group-item
                    >
                  </b-list-group>
                  <h3>price: {{ getPrice(plan.amount) }}</h3>
                  <b-btn
                    @click="
                      (evt) =>
                        purchase(evt, {
                          product: product.id,
                          interval: plan.interval
                        })
                    "
                    :disabled="$store.state.user.plan === plans[2]"
                  >
                    Select
                  </b-btn>
                </b-card-body>
              </b-card>
            </b-col>
          </div>
        </div>
      </b-row>
    </b-container>
  </b-modal>
</template>

<script lang="js">
import Vue from 'vue'
import gql from 'graphql-tag'
import { formatLocaleCurrency } from 'country-currency-map'
import { plans } from '~/assets/config'
export default Vue.extend({
  name: 'Products',
  data() {
    return {
      plans,
      annual: true,
      loading: true
    }
  },
  computed: {
    planInterval() {
      return this.annual ? 'year' : 'month'
    },
    upgrade() {
      return !this.$store.state.auth.user.plan || this.$store.state.auth.user.plan === this.plans[0]
    }
  },
  mounted() {
    this.$store.dispatch('auth/getCountry').then(() => {
      this.loading = false
    }).catch(err => {
      this.$bvToast.toast(err, {
        variant: 'danger',
        title: 'Error'
      })
    })
  },
  methods: {
    show() {
      if (this.$refs['plans-modal']) {
        this.$refs['plans-modal'].show()
      } else {
        this.$bvToast.toast('cannot find plans modal', {
          variant: 'danger',
          title: 'Error'
        })
      }
    },
    getProduct(productName) {
      return this.$store.state.purchase.options.find(option => option.name === productName)
    },
    getPlan(product) {
      if (product) {
        return product.plans.find(option => option.interval === this.planInterval)
      }
      return null
    },
    getPrice(amount) {
      return formatLocaleCurrency(amount * this.$store.state.auth.exchangeRate, this.$store.state.auth.currency)
    },
    cancelSubscription(evt) {
      evt.preventDefault()
      this.$apollo.mutate({mutation: gql`
        mutation cancelSubscription() {
          cancelSubscription() {
            plan
          }
        }
        `, variables: {}})
        .then(({ data }) => {
          this.$store.commit('auth/setPlan', data.cancelSubscription.plan)
          this.$bvToast.toast('cancelled subscription!', {
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
    purchase(evt, item) {
      evt.preventDefault()
      console.log('purchase subscription')
      const err = this.$store.dispatch('purchase/addPlan', item)
      if (err) {
        this.$bvToast.toast(`found error: ${err.message}`, {
          variant: 'danger',
          title: 'Error'
        })
      } else {
        this.$router.push({
          path: '/checkout'
        })
      }
    }
  }
})
</script>

<style lang="scss"></style>

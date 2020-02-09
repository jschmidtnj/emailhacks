<template>
  <b-modal ref="plans-modal" size="xl" title="Plans">
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
        <div :setProductIndex="(firstProductIndex = getProductIndex(plans[0]))">
          <div
            v-if="firstProductIndex >= 0"
            :setPlanIndex="(firstPlanIndex = getPlanIndex(firstProductIndex))"
          >
            <b-col
              v-if="firstProductIndex >= 0 && firstPlanIndex >= 0"
              :setProduct="
                (product =
                  $store.state.purchase.productOptions[firstProductIndex])
              "
              :setPlan="
                (plan =
                  $store.state.purchase.productOptions[firstProductIndex].plans[
                    firstPlanIndex
                  ])
              "
            >
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
                      >max forms: {{ product.maxforms }}</b-list-group-item
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
        <div
          :setProductIndex="(secondProductIndex = getProductIndex(plans[1]))"
        >
          <div
            v-if="secondProductIndex >= 0"
            :setPlanIndex="(secondPlanIndex = getPlanIndex(secondProductIndex))"
          >
            <b-col
              v-if="secondProductIndex >= 0 && secondPlanIndex >= 0"
              :setProduct="
                (product =
                  $store.state.purchase.productOptions[secondProductIndex])
              "
              :setPlan="
                (plan =
                  $store.state.purchase.productOptions[secondProductIndex]
                    .plans[secondPlanIndex])
              "
            >
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
                      >max forms: {{ product.maxforms }}</b-list-group-item
                    >
                  </b-list-group>
                  <h3>price: {{ getPrice(plan.amount) }}</h3>
                  <b-btn
                    @click="(evt) => purchase(evt, plans[1])"
                    :disabled="currentPlan === plans[1]"
                  >
                    Select
                  </b-btn>
                </b-card-body>
              </b-card>
            </b-col>
          </div>
        </div>
        <div :setProductIndex="(thirdProductIndex = getProductIndex(plans[2]))">
          <div
            v-if="thirdProductIndex >= 0"
            :setPlanIndex="(thirdPlanIndex = getPlanIndex(thirdProductIndex))"
          >
            <b-col
              v-if="thirdProductIndex >= 0 && thirdPlanIndex >= 0"
              :setProduct="
                (product =
                  $store.state.purchase.productOptions[thirdProductIndex])
              "
              :setPlan="
                (plan =
                  $store.state.purchase.productOptions[thirdProductIndex].plans[
                    thirdPlanIndex
                  ])
              "
            >
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
                      >max forms: {{ product.maxforms }}</b-list-group-item
                    >
                  </b-list-group>
                  <h3>price: {{ getPrice(plan.amount) }}</h3>
                  <b-btn
                    @click="(evt) => purchase(evt, plans[2])"
                    :disabled="currentPlan === plans[2]"
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
      return !this.currentPlan || this.currentPlan === this.plans[0]
    },
    currentPlan() {
      if (!this.$store.state.auth.user) {
        return null
      }
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
  methods: {
    show() {
      if (this.$refs['plans-modal']) {
        this.$store.dispatch('auth/getCountry').then((res) => {
          this.$store.dispatch('purchase/getProductOptions').then(() => {
            this.loading = false
          }).catch(err => {
            this.$bvToast.toast(err, {
              variant: 'danger',
              title: 'Error'
            })
          })
        }).catch(err => {
          this.$bvToast.toast(err, {
            variant: 'danger',
            title: 'Error'
          })
        })
        this.$refs['plans-modal'].show()
      } else {
        this.$bvToast.toast('cannot find plans modal', {
          variant: 'danger',
          title: 'Error'
        })
      }
    },
    getProductIndex(productName) {
      return this.$store.state.purchase.productOptions.findIndex(option => option.name === productName)
    },
    getPlanIndex(productIndex) {
      if (productIndex >= 0) {
        return this.$store.state.purchase.productOptions[productIndex].plans.findIndex(option => option.interval === this.planInterval)
      }
      return -1
    },
    getPrice(amount) {
      return formatLocaleCurrency(amount * this.$store.state.auth.exchangeRate, this.$store.state.auth.currency)
    },
    cancelSubscription(evt) {
      evt.preventDefault()
      this.$apollo.mutate({mutation: gql`
        mutation cancelSubscription {
          cancelSubscription {
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
    purchase(evt, productName) {
      evt.preventDefault()
      const productIndex = this.getProductIndex(productName)
      const item = {
        productIndex,
        planIndex: this.getPlanIndex(productIndex)
      }
      this.$store.dispatch('purchase/addPlan', item).then((res) => {
        console.log(item)
        if (res) {
          this.$bvToast.toast(`found error: ${res.message}`, {
            variant: 'danger',
            title: 'Error'
          })
        } else {
          console.log('success')
          if (this.$route.path !== '/checkout') {
            this.$router.push({
              path: '/checkout'
            }, () => {
              this.$refs['plans-modal'].hide()
            })
          } else {
            this.$refs['plans-modal'].hide()
          }
        }
      }).catch(err => {
        this.$bvToast.toast(`found error: ${err}`, {
          variant: 'danger',
          title: 'Error'
        })
      })
    }
  }
})
</script>

<style lang="scss"></style>

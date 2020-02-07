<template>
  <b-container>
    <b-card>
      <b-row
        v-if="$store.state.purchase.plan"
        :setProduct="
          (product =
            $store.state.purchase.productOptions[
              $store.state.purchase.plan.productIndex
            ])
        "
      >
        <div
          :setPlan="
            (plan = product.plans[$store.state.purchase.plan.planIndex])
          "
        >
          {{ product.name }}
          {{ formatCurrency(plan.amount) }}
        </div>
      </b-row>
      <b-row
        v-for="(product, index) in $store.state.purchase.products"
        :key="`product-${index}`"
      >
        <b-col>
          {{ product.name }}
          {{ formatCurrency(product.plans[0].amount) }}
        </b-col>
      </b-row>
    </b-card>
  </b-container>
</template>

<script lang="ts">
import Vue from 'vue'
import { formatLocaleCurrency } from 'country-currency-map'
export default Vue.extend({
  name: 'Cart',
  methods: {
    formatCurrency(amount) {
      return formatLocaleCurrency(amount, this.$store.state.auth.currency)
    }
  }
})
</script>

<style lang="scss"></style>

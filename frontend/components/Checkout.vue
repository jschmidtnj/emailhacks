<template>
  <b-card>
    <card
      :class="{ complete }"
      :stripe="stripetoken"
      :options="stripeOptions"
      @change="complete = $event.complete"
      class="stripe-card"
    />
    <button @click="pay" :disabled="!complete" class="pay-with-stripe">
      Pay with credit card
    </button>
  </b-card>
</template>

<script lang="ts">
import Vue from 'vue'
import { Card, createToken } from 'vue-stripe-elements-plus'
export default Vue.extend({
  name: 'Checkout',
  components: {
    Card
  },
  data() {
    const stripetoken = JSON.parse(process.env.stripeconfig).clienttoken
    return {
      complete: false,
      stripetoken,
      stripeOptions: {
        // see https://stripe.com/docs/stripe.js#element-options for details
      }
    }
  },
  methods: {
    pay() {
      // createToken returns a Promise which resolves in a result object with
      // either a token or an error key.
      // See https://stripe.com/docs/api#tokens for the token object.
      // See https://stripe.com/docs/api#errors for the error object.
      // More general https://stripe.com/docs/stripe.js#stripe-create-token.
      createToken().then((data) => console.log(data.token))
    }
  }
})
</script>

<style lang="scss">
.stripe-card {
  width: 300px;
  border: 1px solid grey;
}
.stripe-card.complete {
  border-color: green;
}
</style>

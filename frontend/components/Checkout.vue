<template>
  <b-card>
    <client-only>
      <v-select
        v-model="selectedCountry"
        :options="countryOptions"
        :multiple="false"
        label="countryName"
        aria-describedby="countryfeedback"
      />
    </client-only>
    <card
      :class="{ complete }"
      :stripe="stripetoken"
      :options="stripeOptions"
      @change="complete = $event.complete"
    />
    <b-btn @click="pay" :disabled="!complete" class="mt-4">
      Pay with credit card
    </b-btn>
  </b-card>
</template>

<script lang="ts">
import Vue from 'vue'
import { Card, createSource } from 'vue-stripe-elements-plus'
import { getCodes, getName } from 'country-list'
import flag from 'country-code-emoji'
const stripetoken = JSON.parse(process.env.stripeconfig).clienttoken
const countryOptions = getCodes().map((code) => {
  return {
    countryCode: code,
    countryName: `${flag(code)} ${getName(code)}`
  }
})
// TODO - call payment endpoints, save cart in store
// log in again after payment is complete for different pass
export default Vue.extend({
  name: 'Checkout',
  components: {
    Card
  },
  data() {
    return {
      complete: false,
      stripetoken,
      countryOptions,
      selectedCountry: null,
      userData: {
        address: {
          city: '',
          line1: '',
          line2: '',
          postal_code: '',
          state: ''
        },
        email: '',
        name: '',
        phone: ''
      },
      stripeOptions: {
        // see https://stripe.com/docs/stripe.js#element-options for details
        hidePostalCode: true,
        style: {
          base: {
            fontSize: '18px',
            color: '#32325d',
            fontSmoothing: 'antialiased',
            '::placeholder': {
              color: '#ccc'
            }
          },
          invalid: {
            color: '#e5424d',
            ':focus': {
              color: '#303238'
            }
          }
        }
      }
    }
  },
  mounted() {
    const defaultCountryCode = 'US'
    const setDefaultCountry = (countryCode) => {
      let defaultCountryIndex = countryOptions.findIndex(
        (country) => country.countryCode === countryCode
      )
      if (!defaultCountryIndex) {
        defaultCountryIndex = 0
      }
      this.selectedCountry = countryOptions[defaultCountryIndex]
    }
    this.$axios
      .get('https://www.cloudflare.com/cdn-cgi/trace', {
        baseURL: '',
        headers: null
      })
      .then((res) => {
        if (res.data) {
          const ipData = res.data.split('\n')
          const countryCode = ipData[8].split('=')[1]
          setDefaultCountry(countryCode)
        } else {
          this.$toasted.global.error({
            message: 'cannot get country data'
          })
          setDefaultCountry(defaultCountryCode)
        }
      })
      .catch((err) => {
        this.$toasted.global.error({
          message: err
        })
        setDefaultCountry(defaultCountryCode)
      })
  },
  methods: {
    pay() {
      // createToken returns a Promise which resolves in a result object with
      // either a token or an error key.
      // See https://stripe.com/docs/api#tokens for the token object.
      // See https://stripe.com/docs/api#errors for the error object.
      // More general https://stripe.com/docs/stripe.js#stripe-create-token.
      createSource({
        type: 'card',
        currency: 'usd', // change this to user defined currency
        owner: {
          email: 'jenny.rosen@example.com'
        }
      })
        .then((data) => {
          if (data.errors) {
            this.$toasted.global.error({
              message: `found error(s): ${data.errors}`
            })
          } else if (!data.source && !data.source.id) {
            this.$toasted.global.error({
              message: 'cannot find stripe token'
            })
          } else {
            console.log(data.source)
            console.log(data.source.id)
          }
        })
        .catch((err) => {
          this.$toasted.global.error({
            message: err
          })
        })
    }
  }
})
</script>

<style lang="scss"></style>

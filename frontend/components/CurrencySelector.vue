<template>
  <div>
    <client-only>
      <multiselect
        v-if="!loading"
        v-model="selectedCurrency"
        :options="currencyOptions"
        :allow-empty="false"
        :multiple="false"
        @select="setCurrency"
        label="currencyName"
      />
    </client-only>
  </div>
</template>

<script lang="js">
import Vue from 'vue'
import gql from 'graphql-tag'
import Multiselect from 'vue-multiselect'
import getSymbolFromCurrency from 'currency-symbol-map'
import { defaultCurrency } from '~/assets/config'
export default Vue.extend({
  name: 'CurrencySelector',
  components: {
    Multiselect
  },
  data() {
    return {
      loading: true,
      selectedCurrency: defaultCurrency
    }
  },
  computed: {
    currencyOptions() {
      return this.$store.state.purchase.currencyOptions.map((code) => {
        return {
          currencyCode: code.toLowerCase(),
          currencyName: `${getSymbolFromCurrency(code)} ${code.toUpperCase()}`
        }
      })
    }
  },
  mounted() {
    const init = () => {
      const setDefaultCurrency = (thecurrency) => {
        thecurrency = thecurrency.toLowerCase()
        let defaultCurrencyIndex = this.currencyOptions.findIndex(
          (item) => item.currencyCode === thecurrency
        )
        if (defaultCurrencyIndex < 0) {
          defaultCurrencyIndex = 0
        }
        this.selectedCurrency = this.currencyOptions[defaultCurrencyIndex]
        this.loading = false
      }
      if (this.$store.state.auth.user
        && this.$store.state.auth.user.billing.currency.length > 0) {
        setDefaultCurrency(this.$store.state.auth.user.billing.currency)
      } else {
        setDefaultCurrency(this.$store.state.auth.currency)
      }
    }
    if (this.currencyOptions.length === 0) {
      this.$store.dispatch('purchase/getCurrencyOptions', true).then((res) => {
        console.log('updated currency options')
        init()
      }).catch((err) => {
        this.$bvToast.toast(err.message, {
          variant: 'danger',
          title: 'Error'
        })
      })
    } else {
      init()
    }
  },
  methods: {
    setCurrency(selected) {
      console.log('selected currency')
      const code = selected.currencyCode
      this.$store.commit('auth/setCurrency', code)
      this.$emit('select', code)
      this.$store.dispatch('auth/getExchangeRate', code).then(() => {
        console.log('updated exchange rate')
      }).catch(err => {
        this.$bvToast.toast(err.message, {
          variant: 'danger',
          title: 'Error'
        })
      })
      this.$apollo.mutate({mutation: gql`
        mutation changeBilling($currency: String!) {
          changeBilling(currency: $currency) {
            id
          }
        }
        `, variables: {
          currency: code
        }})
        .then(({ data }) => {
          console.log('updated currency')
        }).catch(err => {
          this.$bvToast.toast(err.message, {
            variant: 'danger',
            title: 'Error'
          })
        })
    }
  }
})
</script>

<style lang="scss"></style>

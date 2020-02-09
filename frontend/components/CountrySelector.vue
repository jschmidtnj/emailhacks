<template>
  <div>
    <client-only>
      <multiselect
        v-if="!loading"
        v-model="selectedCountry"
        :options="countryOptions"
        :multiple="false"
        :allow-empty="false"
        @select="getCountryData"
        label="countryName"
      />
    </client-only>
  </div>
</template>

<script lang="js">
import Vue from 'vue'
import Multiselect from 'vue-multiselect'
import { getName } from 'country-list'
import flag from 'country-code-emoji'
import { defaultCountry } from '~/assets/config'
export default Vue.extend({
  name: 'CountrySelector',
  components: {
    Multiselect
  },
  data() {
    return {
      loading: true,
      countryOptions: [],
      selectedCountry: defaultCountry
    }
  },
  mounted() {
    const setCountryOptions = () => {
      this.countryOptions = this.$store.state.purchase.countryOptions.map((code) => {
        return {
          countryCode: code,
          countryName: `${flag(code)} ${getName(code)}`
        }
      })
      const setDefaultCountry = (thecountry) => {
        let defaultCountryIndex = this.countryOptions.findIndex(
          (country) => country.countryCode === thecountry
        )
        if (defaultCountryIndex < 0) {
          defaultCountryIndex = 0
        }
        this.selectedCountry = this.countryOptions[defaultCountryIndex]
        this.loading = false
      }
      if (this.$store.state.auth.user
        && this.$store.state.auth.user.billing.country.length > 0) {
        setDefaultCountry(this.$store.state.auth.user.billing.country)
      } else {
        setDefaultCountry(this.$store.state.auth.currentCountry)
      }
    }
    if (this.$store.state.purchase.countryOptions.length === 0) {
      this.$store.dispatch('purchase/getCountryOptions').then((res) => {
        setCountryOptions()
      }).catch((err) => {
        this.$bvToast.toast(err.message, {
          variant: 'danger',
          title: 'Error'
        })
      })
    } else {
      setCountryOptions()
    }
  },
  methods: {
    getCountryData(selected) {
      const code = selected.countryCode
      console.log(`updated to selected country: ${code}`)
      this.$emit('select', code)
    }
  }
})
</script>

<style lang="scss"></style>

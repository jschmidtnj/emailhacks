<template>
  <b-container fluid>
    <b-container id="admin-cards">
      <b-row class="my-4">
        <div class="col-lg-6 my-2">
          <section class="card h-100 py-0">
            <div class="card-body">
              <b-form @submit="manageproducts" @reset="resetproduct">
                <span class="card-text">
                  <h2 class="mb-4">{{ mode }} Product</h2>
                  <b-form-group>
                    <label class="form-required">Name</label>
                    <b-input-group>
                      <b-form-input
                        id="name"
                        v-model="product.name"
                        :state="!$v.product.name.$invalid"
                        type="text"
                        class="form-control"
                        aria-describedby="productfeedback"
                        placeholder="name"
                      />
                    </b-input-group>
                    <b-form-invalid-feedback
                      id="productfeedback"
                      :state="!$v.product.name.$invalid"
                    >
                      <div v-if="!$v.product.name.required">
                        name is required
                      </div>
                      <div v-else-if="!$v.product.name.minLength">
                        name must have at least
                        {{ $v.product.name.$params.minLength.min }} characters
                      </div>
                    </b-form-invalid-feedback>
                  </b-form-group>
                  <b-form-group>
                    <label class="form-required">Max Storage</label>
                    <b-form-group append="Bytes">
                      <b-form-input
                        id="maxstorage"
                        v-model="product.maxstorage"
                        :state="!$v.product.maxstorage.$invalid"
                        type="number"
                        class="form-control"
                        aria-describedby="maxstoragefeedback"
                        placeholder="max storage"
                      />
                    </b-form-group>
                    <b-form-invalid-feedback
                      id="maxstoragefeedback"
                      :state="!$v.product.maxstorage.$invalid"
                    >
                      <div v-if="!$v.product.maxstorage.required">
                        max storage is required
                      </div>
                      <div v-else-if="!$v.product.maxstorage.integer">
                        max storage must be an integer
                      </div>
                      <div v-else-if="!$v.product.maxstorage.minValue">
                        max storage must be greater than
                        {{ $v.product.maxstorage.$params.minValue.min }}
                      </div>
                    </b-form-invalid-feedback>
                  </b-form-group>
                  <b-form-group>
                    <label class="form-required">Max Projects</label>
                    <b-form-group>
                      <b-form-input
                        id="maxprojects"
                        v-model="product.maxprojects"
                        :state="!$v.product.maxprojects.$invalid"
                        type="number"
                        class="form-control"
                        aria-describedby="maxprojectsfeedback"
                        placeholder="max projects"
                      />
                    </b-form-group>
                    <b-form-invalid-feedback
                      id="maxprojectsfeedback"
                      :state="!$v.product.maxprojects.$invalid"
                    >
                      <div v-if="!$v.product.maxprojects.required">
                        max projects is required
                      </div>
                      <div v-else-if="!$v.product.maxprojects.integer">
                        max projects must be an integer
                      </div>
                      <div v-else-if="!$v.product.maxprojects.minValue">
                        max projects must be greater than
                        {{ $v.product.maxprojects.$params.minValue.min }}
                      </div>
                    </b-form-invalid-feedback>
                  </b-form-group>
                  <b-form-group>
                    <label class="form-required">Max Forms</label>
                    <b-form-group>
                      <b-form-input
                        id="maxforms"
                        v-model="product.maxforms"
                        :state="!$v.product.maxforms.$invalid"
                        type="number"
                        class="form-control"
                        aria-describedby="maxformsfeedback"
                        placeholder="max forms"
                      />
                    </b-form-group>
                    <b-form-invalid-feedback
                      id="maxformsfeedback"
                      :state="!$v.product.maxforms.$invalid"
                    >
                      <div v-if="!$v.product.maxforms.required">
                        max forms is required
                      </div>
                      <div v-else-if="!$v.product.maxforms.integer">
                        max forms must be an integer
                      </div>
                      <div v-else-if="!$v.product.maxforms.minValue">
                        max forms must be greater than
                        {{ $v.product.maxforms.$params.minValue.min }}
                      </div>
                    </b-form-invalid-feedback>
                  </b-form-group>
                  <h4 class="mt-4">Plans</h4>
                  <div
                    v-for="(planvalue, index) in $v.product.plans.$each.$iter"
                    :key="`plan-${index}`"
                  >
                    <b-form-group class="mb-2">
                      <label class="form-required">Amount</label>
                      <b-form-group prepend="$" append=".00">
                        <b-form-input
                          :id="`plan-amount-${index}`"
                          v-model="product.plans[index].amount"
                          :state="!planvalue.amount.$invalid"
                          type="number"
                          class="form-control"
                          aria-describedby="amountfeedback"
                          placeholder="amount"
                        />
                      </b-form-group>
                      <b-form-invalid-feedback
                        id="amountfeedback"
                        :state="!planvalue.amount.$invalid"
                      >
                        <div v-if="!planvalue.amount.required">
                          amount is required
                        </div>
                        <div v-else-if="!planvalue.amount.integer">
                          amount must be an integer
                        </div>
                        <div v-else-if="!planvalue.amount.minValue">
                          amount must be greater than
                          {{ planvalue.amount.$params.minValue.min }}
                        </div>
                      </b-form-invalid-feedback>
                    </b-form-group>
                    <b-form-group>
                      <label class="form-required">Interval</label>
                      <multiselect
                        :id="`plan-interval-${index}`"
                        v-model="product.plans[index].interval"
                        :options="intervalOptions"
                        :multiple="false"
                        label="interval"
                        aria-describedby="intervalfeedback"
                      />
                      <b-form-invalid-feedback
                        :state="!planvalue.interval.$invalid"
                      >
                        <div v-if="!planvalue.interval.required">
                          interval is required
                        </div>
                        <div v-else-if="!planvalue.interval.validInterval">
                          interval is invalid
                        </div>
                      </b-form-invalid-feedback>
                    </b-form-group>
                  </div>
                  <b-container class="mt-4">
                    <b-row>
                      <b-col>
                        <b-btn @click="addPlan" variant="primary" class="mr-2">
                          <client-only>
                            <font-awesome-icon
                              class="mr-2 arrow-size-edit"
                              icon="plus-circle"
                            /> </client-only
                          >Add
                        </b-btn>
                        <b-btn
                          :disabled="product.plans.length === 1"
                          @click="removePlan"
                          variant="primary"
                          class="mr-2"
                        >
                          <client-only>
                            <font-awesome-icon
                              class="mr-2 arrow-size-edit"
                              icon="times"
                            /> </client-only
                          >Remove
                        </b-btn>
                      </b-col>
                    </b-row>
                  </b-container>
                  <b-container class="mt-4">
                    <b-row>
                      <b-col>
                        <b-btn
                          :disabled="$v.product.$invalid || submitting"
                          variant="primary"
                          type="submit"
                        >
                          <client-only>
                            <font-awesome-icon
                              class="mr-2 arrow-size-edit"
                              icon="angle-double-right"
                            /> </client-only
                          >Submit
                        </b-btn>
                        <b-btn variant="primary" type="reset" class="mr-4">
                          <client-only>
                            <font-awesome-icon
                              class="mr-2 arrow-size-edit"
                              icon="times"
                            /> </client-only
                          >Clear
                        </b-btn>
                      </b-col>
                    </b-row>
                  </b-container>
                </span>
              </b-form>
            </div>
          </section>
        </div>
        <div class="col-lg-6 my-2">
          <section class="card h-100 py-0">
            <div class="card-body">
              <b-form @submit="searchproducts" @reset="clearsearch">
                <span class="card-text">
                  <h2 class="mb-4">Currencies</h2>
                  <client-only>
                    <multiselect
                      v-if="!loading"
                      v-model="selectedCurrencies"
                      :options="currencyOptions"
                      :multiple="true"
                      @select="addCurrency"
                      @remove="removeCurrency"
                      label="currencyName"
                      track-by="currencyCode"
                      class="mb-2"
                    />
                  </client-only>
                  <h2 class="mb-4">Search</h2>
                  <b-form-group>
                    <label class="form-required">Query</label>
                    <span>
                      <b-form-input
                        v-model="search"
                        :state="!$v.search.$invalid"
                        type="text"
                        class="form-control mb-2"
                        aria-describedby="searchfeedback"
                        placeholder="search..."
                      />
                    </span>
                    <b-form-invalid-feedback
                      id="searchfeedback"
                      :state="!$v.search.$invalid"
                    >
                      <div v-if="!$v.search.required">query is required</div>
                      <div v-else-if="!$v.search.minLength">
                        query must have at least
                        {{ $v.search.$params.minLength.min }} characters
                      </div>
                    </b-form-invalid-feedback>
                  </b-form-group>
                  <b-btn
                    :disabled="$v.search.$invalid"
                    variant="primary"
                    type="submit"
                    class="mr-4"
                  >
                    <client-only>
                      <font-awesome-icon
                        class="mr-2 arrow-size-edit"
                        icon="angle-double-right"
                      /> </client-only
                    >Search
                  </b-btn>
                  <b-btn variant="primary" type="reset" class="mr-4">
                    <client-only>
                      <font-awesome-icon
                        class="mr-2 arrow-size-edit"
                        icon="times"
                      /> </client-only
                    >Clear
                  </b-btn>
                  <br />
                  <br />
                </span>
              </b-form>
              <b-table
                :items="searchresults"
                :fields="fields"
                show-empty
                stacked="md"
              >
                <template v-slot:cell(name)="data">
                  {{ data.value }}
                </template>
                <template v-slot:cell(created)="data">
                  {{ formatDate(data.value, 'M/d/yyyy') }}
                </template>
                <template v-slot:cell(id)="data">
                  {{ data.value }}
                </template>
                <template v-slot:cell(actions)="data">
                  <b-button
                    @click="editProduct(data.item)"
                    size="sm"
                    class="mr-1"
                  >
                    Edit
                  </b-button>
                  <b-button @click="deleteProduct(data.item)" size="sm">
                    Del
                  </b-button>
                </template>
              </b-table>
            </div>
          </section>
        </div>
      </b-row>
    </b-container>
  </b-container>
</template>

<script lang="js">
import Vue from 'vue'
import gql from 'graphql-tag'
import Multiselect from 'vue-multiselect'
import getSymbolFromCurrency from 'currency-symbol-map'
import { validationMixin } from 'vuelidate'
import { required, minLength, minValue, integer } from 'vuelidate/lib/validators'
import { formatRelative } from 'date-fns'
import { clone } from '~/assets/utils'
// @ts-ignore
const seo = JSON.parse(process.env.seoconfig)
const intervalOptions = [
  {
    label: 'One-time Purchase',
    interval: 'once'
  },
  {
    label: 'Monthly',
    interval: 'month'
  },
  {
    label: 'Yearly',
    interval: 'year'
  }
]
const validInterval = (selected) => selected && intervalOptions.some(curr => curr.interval === selected.interval)
/**
 * products edit
 */
const modetypes = {
  add: 'Add',
  edit: 'Edit',
  delete: 'Delete'
}
const defaultPlan = {
  interval: 'month',
  amount: 0
}
const defaultProduct = {
  name: '',
  plans: [
    clone(defaultPlan)
  ],
  maxprojects: 1,
  maxforms: 1,
  maxstorage: 1
}
export default Vue.extend({
  name: 'ProductEdit',
  // @ts-ignore
  layout: 'admin',
  components: {
    Multiselect
  },
  mixins: [validationMixin],
  // @ts-ignore
  data() {
    return {
      submitting: false,
      loading: true,
      selectedCurrencies: [],
      currencyOptions: [],
      modetypes,
      mode: modetypes.add,
      productid: null,
      search: '',
      type: 'product',
      searchresults: [],
      intervalOptions,
      fields: [
        {
          key: 'name',
          label: 'Name',
          sortable: true
        },
        {
          key: 'created',
          label: 'Created',
          sortable: true
        },
        {
          key: 'id',
          label: 'ID',
          sortable: true
        },
        {
          key: 'actions',
          label: 'Actions',
          sortable: false
        }
      ],
      product: clone(defaultProduct)
    }
  },
  // @ts-ignore
  validations: {
    search: {
      required,
      minLength: minLength(3)
    },
    product: {
      name: {
        required,
        minLength: minLength(3)
      },
      maxprojects: {
        required,
        integer,
        minValue: minValue(0)
      },
      maxstorage: {
        required,
        integer,
        minValue: minValue(0)
      },
      maxforms: {
        required,
        integer,
        minValue: minValue(0)
      },
      plans: {
        $each: {
          interval: {
            required,
            validInterval
          },
          amount: {
            required,
            integer,
            minValue: minValue(0)
          }
        }
      }
    }
  },
  // @ts-ignore
  head() {
    const title = 'Admin Edit Product'
    const description = 'admin page for editing products'
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
  mounted() {
    this.$apollo.query({query: gql`
      query currencyOptions {
        currencyOptions
      }
      `, variables: {},
      fetchPolicy: 'network-only'
      })
      .then(({ data }) => {
        if (!data.currencyOptions) {
          this.$bvToast.toast('cannot find currency options', {
            variant: 'danger',
            title: 'Error'
          })
        } else {
          this.currencyOptions = data.currencyOptions.map((code) => {
            return {
              currencyCode: code.toLowerCase(),
              currencyName: `${getSymbolFromCurrency(code)} ${code.toUpperCase()}`
            }
          })
          const setSelectedCurrencies = () => {
            this.selectedCurrencies = this.$store.state.purchase.currencyOptions.map(currencyCode =>
              this.currencyOptions.find(elem => elem.currencyCode === currencyCode)
            )
            this.loading = false
          }
          if (this.$store.state.purchase.currencyOptions.length === 0) {
            this.$store.dispatch('purchase/getCurrencyOptions', false).then((res) => {
              console.log('updated currency options')
              setSelectedCurrencies()
            }).catch((err) => {
              this.$bvToast.toast(err.message, {
                variant: 'danger',
                title: 'Error'
              })
            })
          } else {
            setSelectedCurrencies()
          }
        }
      }).catch(err => {
        this.$bvToast.toast(err.message, {
          variant: 'danger',
          title: 'Error'
        })
      })
  },
  /* eslint-disable */
  methods: {
    mongoidToDate(id) {
      return parseInt(id.substring(0, 8), 16) * 1000
    },
    addCurrency(selectedOption) {
      console.log('add currency')
      const currency = selectedOption.currencyCode
      this.$apollo.mutate({
        mutation: gql`
          mutation addCurrency($currency: String!) {
            addCurrency(currency: $currency)
          }
        `, variables: {
          currency
        }
      })
      .then(({ data }) => {
        this.$store.commit('purchase/addCurrencyOption', currency)
        this.$bvToast.toast('added currency', {
          variant: 'success',
          title: 'Success'
        })
      })
      .catch((err) => {
        console.error(err)
        this.$bvToast.toast(`found error: ${err.message}`, {
          variant: 'danger',
          title: 'Error'
        })
        this.selectedCurrencies = this.selectedCurrencies.slice(selectedOption, 1)
      })
    },
    removeCurrency(selectedOption) {
      const currency = selectedOption.currencyCode
      this.$apollo.mutate({
        mutation: gql`
          mutation deleteCurrency($currency: String!) {
            deleteCurrency(currency: $currency)
          }
        `, variables: {
          currency
        }
      })
      .then(({ data }) => {
        this.$store.commit('purchase/removeCurrencyOption', currency)
        this.$bvToast.toast('removed currency', {
          variant: 'success',
          title: 'Success'
        })
      })
      .catch((err) => {
        console.error(err)
        this.$bvToast.toast(`found error: ${err.message}`, {
          variant: 'danger',
          title: 'Error'
        })
        this.selectedCurrencies.push(selectedOption)
      })
    },
    editProduct(searchresult) {
      this.productid = searchresult.id
      // get product data first
      this.$apollo.query({
        query: gql`
          query product($id: String!) {
            product(id: $id) {
              name
              maxprojects
              maxforms
              maxstorage
              plans {
                amount
                interval
              }
            }
          }`,
          variables: {id: this.productid},
          fetchPolicy: 'network-only'
        }).then(({ data }) => {
          this.mode = this.modetypes.edit
          data.product.plans.map(plan => plan.interval = intervalOptions.find(planOptions => planOptions.interval === plan.interval))
          this.product = data.product
          this.$bvToast.toast(`edit product with id ${this.productid}`, {
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
    addPlan(evt) {
      evt.preventDefault()
      this.product.plans.push(clone(defaultPlan))
    },
    removePlan(evt) {
      evt.preventDefault()
      this.product.plans.pop()
    },
    deleteProduct(searchresult) {
      const id = searchresult.id
      this.$apollo.mutate({mutation: gql`
        mutation deleteProduct($id: String!){deleteProduct(id: $id){id} }
        `, variables: {id: id}})
        .then(({ data }) => {
          this.searchresults.splice(
            this.searchresults.indexOf(searchresult),
            1
          )
          this.$bvToast.toast('product deleted', {
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
    formatDate(dateUTC) {
      return formatRelative(dateUTC, new Date())
    },
    searchproducts(evt) {
      evt.preventDefault()
      this.$apollo.query({
        query: gql`
          query products {
            products {
              name,
              id
            }
          }`,
          variables: {},
          fetchPolicy: 'network-only'
        }).then(({ data }) => {
          const products = data.products
          products.map(
            product => {
              product.created = this.mongoidToDate(product.id)
            }
          )
          this.searchresults = products
          this.$bvToast.toast(`found ${this.searchresults.length} result${
              this.searchresults.length === 1 ? '' : 's'
            }`, {
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
    clearsearch(evt) {
      if (evt) evt.preventDefault()
      this.search = ''
      this.searchresults = []
    },
    resetproduct(evt) {
      if (evt) evt.preventDefault()
      this.product = clone(defaultProduct)
      this.mode = this.modetypes.add
      this.productid = null
    },
    manageproducts(evt) {
      evt.preventDefault()
      let productid = this.productid
      this.submitting = true
      const onSuccess = () => {
        this.$bvToast.toast(`${this.mode}ed product with id ${productid}`, {
          variant: 'success',
          title: 'Sucess'
        })
        this.submitting = false
        this.resetproduct(null)
      }
      const plans = clone(this.product.plans)
      plans.map(plan => plan.interval = plan.interval.interval)
      console.log(plans)
      if (this.mode === this.modetypes.add) {
        this.$apollo.mutate({mutation: gql`
          mutation addProduct($name: String!, $maxprojects: Int!, $maxstorage: Int!, $maxforms: Int!, $plans: [PlanInput!]!) {
            addProduct(name: $name, maxprojects: $maxprojects, maxstorage: $maxstorage, maxforms: $maxforms, plans: $plans) {
              id
            }
          }`, variables: {name: this.product.name, maxstorage: this.product.maxstorage, maxprojects: this.product.maxstorage, maxforms: this.product.maxforms, plans}})
          .then(({ data }) => {
            productid = data.addProduct.id
            onSuccess()
          }).catch(err => {
            console.error(err)
            this.$bvToast.toast(`found error: ${err.message}`, {
              variant: 'danger',
              title: 'Error'
            })
          })
      } else {
        this.$apollo.mutate({mutation: gql`
          mutation updateProduct($id: String!, $name: String!, $maxprojects: Int!, $maxstorage: Int!, $maxforms: Int!, $plans: [PlanInput!]!)
          {updateProduct(id: $id, name: $name, maxprojects: $maxprojects, maxstorage: $maxstorage, maxforms: $maxforms, plans: $plans){id} }
          `, variables: {id: this.productid, name: this.product.name, maxstorage: this.product.maxstorage, maxprojects: this.product.maxstorage, maxforms: this.product.maxforms, plans}})
          .then(({ data }) => {
            onSuccess()
          }).catch(err => {
            console.error(err)
            this.$bvToast.toast(`found error: ${err.message}`, {
              variant: 'danger',
              title: 'Error'
            })
          })
      }
    }
  }
})
</script>

<style lang="scss"></style>

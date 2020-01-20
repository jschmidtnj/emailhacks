<template>
  <div>
    <div id="blog-data" v-if="blog">
      <div id="blog-content">
        <b-container class="hero-body">
          <b-row>
            <b-col>
              <b-img-lazy
                v-if="blog.heroimage"
                :blank-src="
                  `${blogCdn}/${staticStorageIndexes.blogfiles}/${
                    blog.id
                  }/${blog.heroimage.id + paths.blur}`
                "
                :src="
                  `${blogCdn}/${staticStorageIndexes.blogfiles}/${
                    blog.id
                  }/${blog.heroimage.id + paths.original}`
                "
                :alt="blog.heroimage.name"
                class="hero-img m-0"
              />
              <div class="main-overlay">
                <div class="text-overlay">
                  <!-- add text overlay here -->
                </div>
              </div>
            </b-col>
          </b-row>
        </b-container>
        <b-container id="header-container" v-if="blog">
          <h1>{{ blog.title }}</h1>
          <p>{{ blog.author }}</p>
          <p v-if="blog.id">
            {{ formatDate(mongoidToDate(blog.id), 'M/d/yyyy') }}
          </p>
          <p>{{ blog.views }}</p>
          <a :href="`${shortlinkurl}/${blog.shortlink}`">
            {{ `${shortlinkurl}/${blog.shortlink}` }}
          </a>
          <p class="orange-text">
            {{ blog.categories.join(' | ') }}
          </p>
          <hr />
        </b-container>
        <b-container id="content-container" v-if="blog">
          <vue-markdown
            :source="blog.content"
            @rendered="updateMarkdown"
            class="markdown"
          />
        </b-container>
      </div>
    </div>
    <page-loading v-else :loading="true" />
  </div>
</template>

<script lang="js">
import Vue from 'vue'
import { format } from 'date-fns'
import VueMarkdown from 'vue-markdown'
import Prism from 'prismjs'
import LazyLoad from 'vanilla-lazyload'
import PageLoading from '~/components/PageLoading.vue'
import {
  cloudStorageURLs,
  staticStorageIndexes,
  paths,
  adminTypes
} from '~/assets/config'
const lazyLoadInstance = new LazyLoad({
  elements_selector: '.lazy'
})
// @ts-ignore
const ampurl = process.env.ampurl
// @ts-ignore
const seo = JSON.parse(process.env.seoconfig)
// @ts-ignore
const shortlinkurl = process.env.shortlinkurl
export default Vue.extend({
  name: 'Blog',
  components: {
    VueMarkdown,
    PageLoading
  },
  data() {
    return {
      id: null,
      type: 'blog',
      blog: null,
      shortlinkurl,
      blogCdn: cloudStorageURLs.static,
      staticStorageIndexes,
      paths
    }
  },
  /* eslint-disable */
  mounted() {
    if (this.$route.params && this.$route.params.id) {
      this.id = this.$route.params.id
      const useCache = this.$store.state.auth.user && adminTypes.includes(this.$store.state.auth.user.type)
      this.$apollo.query({
        query: gql`
          query blog($id: String!, $cache: Boolean!) {
            blog(id: $id, cache: $cache) {
              title,
              caption,
              content,
              id,
              author,
              views,
              shortlink,
              heroimage{
                name,
                id
              },
              tileimage{
                id
              },
              categories,
              tags
            }
          }`,
          variables: {id: this.id, cache: useCache},
          fetchPolicy: useCache ? 'cache-first' : 'network-only'
        }).then(({ data }) => {
          const blog = data.blog
          this.blog = blog
          // update title for spa
          document.title = this.blog.title
        }).catch(err => {
          console.error(err)
          this.$toasted.global.error({
            message: `found error: ${err.message}`
          })
        })
    } else {
      this.$nuxt.error({
        statusCode: 404,
        message: 'could not find id in params'
      })
    }
  },
  // @ts-ignore
  head() {
    const title = this.blog ? this.blog.title : 'Blog'
    const description = this.blog ? this.blog.caption : 'Blog'
    const meta = [
      { property: 'og:title', content: title },
      { property: 'og:description', content: description },
      { name: 'twitter:title', content: title },
      {
        name: 'twitter:description',
        content: description
      },
      { hid: 'description', name: 'description', content: description }
    ]
    const script = []
    if (this.blog) {
      const image = `${cloudStorageURLs.static}/${this.staticStorageIndexes.blogfiles
      }/${this.blog.id}/${this.blog.tileimage.id + this.paths.original}`
      meta.push({
        property: 'og:image',
        content: image
      })
      meta.push({
        name: 'twitter:image',
        content: image
      })
      const date = this.formatDate(
        this.mongoidToDate(this.blog.id),
        'YYYY-MM-DD'
      )
      script.push({
        innerHTML: JSON.stringify({
          '@context': 'https://schema.org',
          '@type': 'BlogPosting',
          headline: this.blog.title,
          alternativeHeadline: this.blog.caption,
          image: image,
          editor: this.blog.author,
          genre: this.blog.categories.join(' '),
          keywords: this.blog.tags.join(' '),
          wordcount: this.blog.content.length,
          publisher: seo.url,
          url: seo.url,
          datePublished: date,
          dateCreated: date,
          dateModified: date,
          description: this.blog.caption,
          articleBody: this.blog.content,
          author: {
            '@type': 'Person',
            name: this.blog.author
          }
        }),
        type: 'application/ld+json'
      })
    }
    return {
      title: title,
      meta: meta,
      link: [
        {
          rel: 'amphtml',
          href: `${ampurl}/blog/${this.$route.query.id}`
        }
      ],
      __dangerouslyDisableSanitizers: ['script'],
      script: script
    }
  },
  methods: {
    updateMarkdown() {
      this.$nextTick(() => {
        Prism.highlightAll()
        if (lazyLoadInstance) {
          console.log('update lazyload')
          lazyLoadInstance.update()
        }
      })
    },
    formatDate(dateUTC, formatStr) {
      return format(dateUTC, formatStr)
    },
    mongoidToDate(id) {
      return parseInt(id.substring(0, 8), 16) * 1000
    }
  }
})
</script>

<style lang="scss">
@import '~/node_modules/prismjs/themes/prism.css';
#content-container {
  padding-left: 0;
  padding-right: 0;
}
#content-container p,
h1,
h2,
h3,
h4,
h5,
h6 {
  padding-right: 15px;
  padding-left: 15px;
}
#blog-data {
  display: flex;
  min-height: 90vh;
  flex-direction: column;
}
#blog-content {
  flex: 1;
}
@media (min-width: 1200px) {
  .container {
    max-width: 1400px;
  }
}
.white-color {
  color: white;
}
.hero-img {
  object-fit: cover;
  width: 100%;
  // set max height for image
  max-height: 40em;
  position: relative;
}
.hero-body {
  overflow: hidden;
  text-align: center;
  width: 100%;
  // set max height for image
  max-height: 40em;
  padding: 0;
}
.main-overlay {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  z-index: 9999;
  // add gradiant to show text clearly
  // background: linear-gradient(rgba(0, 0, 0, 0.2), rgba(0, 0, 0, 0.2));
}
.text-overlay {
  padding-top: 10%;
  height: 100%;
}
</style>

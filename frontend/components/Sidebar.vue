<template>
  <div>
    <nav id="sidebar">
      <ul class="nav flex-column components">
        <li class="nav-item">
          <div class="sidebar-header">
            <nuxt-link to="/" class="no-underline">
              <h3>Mail Pear</h3>
            </nuxt-link>
          </div>
        </li>
        <li class="nav-item">
          <a @click="navigationPath" href="#" class="nav-link">
            {{ inProject ? 'Back' : 'All Projects' }}
          </a>
        </li>
        <li class="nav-item">
          <nuxt-link v-if="!inProject" to="/project" class="no-underline">
            <a class="nav-link">New Project</a>
          </nuxt-link>
        </li>
        <li v-if="inProject && !inForm" class="nav-item">
          <nuxt-link :to="`/project/${projectId}/form`" class="no-underline">
            <a class="nav-link">New Form</a>
          </nuxt-link>
        </li>
      </ul>
    </nav>
  </div>
</template>

<script lang="js">
import Vue from 'vue'
export default Vue.extend({
  name: 'Sidebar',
  data() {
    return {
      projectPath: '/project/',
      formPath: '/form/'
    }
  },
  computed: {
    inProject() {
      return this.$nuxt.$route.path.includes(this.projectPath)
    },
    inForm() {
      return this.$nuxt.$route.path.includes(this.formPath)
    },
    projectId() {
      const projectPathIndex = this.$route.path.indexOf(this.projectPath)
      if (projectPathIndex > 0)
        return null
      const after = this.$route.path.substring(projectPathIndex + this.projectPath.length)
      const extraIndex = this.$route.path.indexOf('/')
      if (extraIndex > 0) {
        return after.substring(0, extraIndex)
      }
      return after
    }
  },
  methods: {
    navigationPath(evt) {
      evt.preventDefault()
      if (this.inProject) {
        this.$router.go(-1)
      } else {
        this.$router.push({ path: '/projects' })
      }
    }
  }
})
</script>

<style lang="scss" scoped>
#sidebar {
  width: 250px;
  position: fixed;
  top: 0;
  left: 0;
  height: 100vh;
  background: #390066;
  color: white;
  transition: all 0.3s;
}

#sidebar .sidebar-header {
  padding-left: 15px;
  padding-top: 20px;
  padding-bottom: 20px;
}
#sidebar .sidebar-header :hover {
  color: white;
}

#sidebar ul li a {
  color: white;
}

#sidebar ul li a:hover {
  color: #35406d;
  background: white;
}

@media (max-width: 768px) {
  #sidebar {
    margin-left: -250px;
  }
  #sidebar.active {
    margin-left: 0;
  }
  .main-wrapper-sidebar {
    width: 100%;
  }
  .main-wrapper-sidebar.active {
    width: calc(100% - 250px);
  }
  #sidebarCollapse span {
    display: none;
  }
}
</style>

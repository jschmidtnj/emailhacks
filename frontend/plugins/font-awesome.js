import Vue from 'vue'
import { library } from '@fortawesome/fontawesome-svg-core'
import {
  faGripHorizontal,
  faTimes,
  faTrash,
  faBold,
  faItalic,
  faStrikethrough,
  faUnderline,
  faCode,
  faParagraph,
  faListOl,
  faListUl,
  faQuoteRight,
  faGripLines,
  faUndo,
  faRedo,
  faPlus,
  faPaperPlane,
  faImage,
  faPlusCircle,
  faAngleDoubleRight,
  faSmile,
  faShare,
  faShoppingCart
} from '@fortawesome/free-solid-svg-icons'

import {} from '@fortawesome/free-brands-svg-icons'

import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'

library.add(
  faGripHorizontal,
  faTimes,
  faTrash,
  faBold,
  faItalic,
  faStrikethrough,
  faUnderline,
  faCode,
  faParagraph,
  faListOl,
  faListUl,
  faQuoteRight,
  faGripLines,
  faUndo,
  faRedo,
  faPlus,
  faPaperPlane,
  faImage,
  faPlusCircle,
  faAngleDoubleRight,
  faSmile,
  faShare,
  faShoppingCart
)

Vue.component('font-awesome-icon', FontAwesomeIcon)

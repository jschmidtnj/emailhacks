import Fuse from 'fuse.js'
import { Node } from 'tiptap'
import { replaceText } from 'tiptap-commands'
import SuggestionsPlugin from 'tiptap-extensions/src/plugins/Suggestions'

export default class Emoji extends Node {
  get name() {
    return 'emoji'
  }
  get defaultOptions() {
    return {
      matcher: {
        char: ':',
        allowSpaces: false,
        startOfLine: false
      },
      emojiClass: 'emoji',
      suggestionClass: 'emoji-suggestion',
      onFilter: (items, query) => {
        if (!query) {
          return items
        }
        const fuse = new Fuse(items, {
          threshold: 0.2,
          keys: ['id']
        })
        return fuse.search(query)
      }
    }
  }

  get schema() {
    return {
      attrs: {
        id: {},
        label: {}
      },
      group: 'inline',
      inline: true,
      selectable: false,
      atom: true,
      toDOM: (node) => [
        'span',
        {
          class: this.options.emojiClass,
          'data-emoji-id': node.attrs.id
        },
        node.attrs.label
      ],
      parseDOM: [
        {
          tag: 'span[data-emoji-id]',
          getAttrs: (dom) => {
            const id = dom.getAttribute('data-emoji-id')
            const label = dom.textContent
              .split(this.options.matcher.char)
              .join('')
            return { id, label }
          }
        }
      ]
    }
  }

  commands({ schema }) {
    return (attrs) => replaceText(null, schema.nodes[this.name], attrs)
  }

  get plugins() {
    return [
      SuggestionsPlugin({
        command: ({ range, attrs, schema }) =>
          replaceText(range, schema.nodes[this.name], attrs),
        appendText: ' ',
        matcher: this.options.matcher,
        items: this.options.items,
        onEnter: this.options.onEnter,
        onChange: this.options.onChange,
        onExit: this.options.onExit,
        onKeyDown: this.options.onKeyDown,
        onFilter: this.options.onFilter,
        suggestionClass: this.options.suggestionClass
      })
    ]
  }
}

import elasticsearch = require('elasticsearch')
import { elasticuri } from './config'

/**
 * elastic functions - initialize elasticsearch
 */

const fileMappings = {
  id: {
    type: 'keyword'
  },
  name: {
    type: 'keyword'
  },
  width: {
    type: 'integer'
  },
  height: {
    type: 'integer'
  },
  type: {
    type: 'keyword'
  }
}

export const blogMappings = {
  properties: {
    title: {
      type: 'keyword'
    },
    caption: {
      type: 'keyword'
    },
    author: {
      type: 'keyword'
    },
    color: {
      type: 'text'
    },
    tags: {
      type: 'keyword'
    },
    categories: {
      type: 'keyword'
    },
    content: {
      type: 'text'
    },
    views: {
      type: 'integer'
    },
    created: {
      type: 'date',
      format: 'epoch_millis'
    },
    updated: {
      type: 'date',
      format: 'epoch_millis'
    },
    heroimage: {
      type: 'nested',
      properties: fileMappings
    },
    tileimage: {
      type: 'nested',
      properties: fileMappings
    },
    files: {
      type: 'nested',
      properties: fileMappings
    },
    comments: {
      type: 'nested'
    }
  }
}

const linkAccessMappings = {
  shortlink: {
    type: 'keyword'
  },
  secret: {
    type: 'keyword'
  },
  type: {
    type: 'keyword'
  }
}

export const formMappings = {
  properties: {
    name: {
      type: 'text'
    },
    items: {
      type: 'nested'
    },
    owner: {
      type: 'keyword'
    },
    multiple: {
      type: 'boolean'
    },
    access: {
      type: 'object'
    },
    linkaccess: {
      type: 'object',
      properties: linkAccessMappings
    },
    views: {
      type: 'integer'
    },
    responses: {
      type: 'integer'
    },
    public: {
      type: 'keyword'
    },
    files: {
      type: 'nested',
      properties: fileMappings
    },
    created: {
      type: 'date',
      format: 'epoch_millis'
    },
    updated: {
      type: 'date',
      format: 'epoch_millis'
    },
    project: {
      type: 'keyword'
    }
  }
}

export const projectMappings = {
  properties: {
    name: {
      type: 'text'
    },
    forms: {
      type: 'integer'
    },
    owner: {
      type: 'keyword'
    },
    access: {
      type: 'object'
    },
    linkaccess: {
      type: 'object',
      properties: linkAccessMappings
    },
    views: {
      type: 'integer'
    },
    public: {
      type: 'keyword'
    },
    created: {
      type: 'date',
      format: 'epoch_millis'
    },
    updated: {
      type: 'date',
      format: 'epoch_millis'
    }
  }
}

export const responseMappings = {
  properties: {
    views: {
      type: 'integer'
    },
    user: {
      type: 'keyword'
    },
    owner: {
      type: 'keyword'
    },
    form: {
      type: 'keyword'
    },
    project: {
      type: 'keyword'
    },
    created: {
      type: 'date',
      format: 'epoch_millis'
    },
    updated: {
      type: 'date',
      format: 'epoch_millis'
    },
    items: {
      type: 'nested'
    },
    files: {
      type: 'nested',
      properties: fileMappings
    }
  }
}

const indexsettings = {
  number_of_shards: 1,
  number_of_replicas: 0
}

const writeclient = new elasticsearch.Client({
  host: elasticuri
})

export const initializeElasticMappings = (indexname, doctype, mappings) => {
  return new Promise((resolve, reject) => {
    writeclient
      .ping({
        requestTimeout: 1000
      })
      .then(() => {
        console.log(`able to ping writeclient`)
        writeclient.indices
          .putSettings({
            index: indexname,
            body: {
              index: indexsettings
            }
          })
          .then(res0 => {
            console.log(
              `deleted all shards in ${indexname}: ${JSON.stringify(
                res0
              )}`
            )
          })
          .catch(err => {
            const errmessage = `error deleting shards in index ${indexname}: ${err}`
            console.log(errmessage)
          })
          .then(() => {
            writeclient.indices
              .delete({
                index: indexname
              })
              .then(res1 => {
                console.log(
                  `deleted index ${indexname}: ${JSON.stringify(res1)}`
                )
              })
              .catch(err => {
                const errmessage = `error deleting index ${indexname}: ${err}`
                console.log(errmessage)
              })
              .then(() => {
                return writeclient.indices
                  .exists({
                    index: indexname
                  })
                  .then(res2 => {
                    if (res2) {
                      console.log(`index ${indexname} exists still`)
                    } else {
                      return writeclient.indices
                        .create({
                          index: indexname,
                          body: {
                            settings: indexsettings
                          }
                        })
                        .then(res3 => {
                          console.log(`added index ${indexname}: ${res3}`)
                          return writeclient.indices
                            .getMapping()
                            .then(res4 => {
                              if (
                                Object.keys(res4[indexname].mappings)
                                  .length === 0
                              ) {
                                console.log(
                                  `${indexname}: no mappings :)`
                                )
                                return writeclient.indices
                                  .putMapping({
                                    index: indexname,
                                    type: doctype,
                                    body: mappings,
                                    include_type_name: true
                                  })
                                  .then(res5 => {
                                    console.log(
                                      `initialized ${indexname}: ${JSON.stringify(
                                        res5
                                      )}`
                                    )
                                    resolve(`finished initializing elasticsearch`)
                                  })
                                  .catch(err => {
                                    const errmessage = `could not create mapping for ${indexname}: ${err}`
                                    console.log(errmessage)
                                    reject(errmessage)
                                  })
                              } else {
                                const errmessage = `${indexname} already has mappings :(`
                                console.log(errmessage)
                                reject(errmessage)
                              }
                            })
                            .catch(err => {
                              const errmessage = `could not get mappings for ${indexname}: ${err}`
                              console.log(errmessage)
                              reject(errmessage)
                            })
                        })

                        .catch(err => {
                          const errmessage = `error adding index ${indexname}: ${err}`
                          console.log(errmessage)
                          reject(errmessage)
                        })
                    }
                  })
                  .catch(err => {
                    const errmessage = `error checking if index ${indexname} exists: ${err}`
                    console.log(errmessage)
                    reject(errmessage)
                  })
              })
          })
      })
      .catch(err => {
        const errmessage = `unable to ping writeclient`
        console.log(errmessage)
        reject(err)
      })
  })
}

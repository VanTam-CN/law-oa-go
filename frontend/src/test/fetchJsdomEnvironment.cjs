const { TestEnvironment } = require('jest-environment-jsdom')

// The data router constructs browser Requests even without loaders or network
// traffic. Use Node's built-in Fetch classes; fetch itself stays test-mocked.
module.exports = class FetchJsdomEnvironment extends TestEnvironment {
  async setup() {
    await super.setup()
    this.global.Request = Request
    this.global.Response = Response
    this.global.Headers = Headers
  }
}

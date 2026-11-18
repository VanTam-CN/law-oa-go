import {
  buildOnlyOfficeScriptUrl,
  encodeJsonForScript,
  getOnlyOfficeAllowedOrigins,
  isOnlyOfficeEnabledFromEnv,
  normalizeOnlyOfficeOrigin,
} from '../onlyoffice'

describe('OnlyOffice config helpers', () => {
  it('normalizes only plain HTTP(S) origins', () => {
    expect(normalizeOnlyOfficeOrigin('https://docs.example.com')).toBe('https://docs.example.com')
    expect(normalizeOnlyOfficeOrigin('http://docs.example.com:8080/')).toBe(
      'http://docs.example.com:8080',
    )
  })

  it('rejects paths, queries, fragments, credentials, and unsupported schemes', () => {
    expect(normalizeOnlyOfficeOrigin('https://docs.example.com/editor')).toBeNull()
    expect(normalizeOnlyOfficeOrigin('https://docs.example.com?token=1')).toBeNull()
    expect(normalizeOnlyOfficeOrigin('https://docs.example.com/#editor')).toBeNull()
    expect(normalizeOnlyOfficeOrigin('https://user:pass@docs.example.com')).toBeNull()
    expect(normalizeOnlyOfficeOrigin('ftp://docs.example.com')).toBeNull()
  })

  it('builds the editor script URL from the normalized origin', () => {
    expect(buildOnlyOfficeScriptUrl('https://docs.example.com')).toBe(
      'https://docs.example.com/web-apps/apps/api/documents/api.js',
    )
  })

  it('encodes JSON so server-controlled values cannot close the script context', () => {
    const encoded = encodeJsonForScript({
      name: '</script><script>alert("xss")</script>',
      url: 'https://docs.example.com/a&b',
    })

    expect(encoded).not.toContain('</script>')
    expect(encoded).not.toContain('<script>')
    expect(JSON.parse(encoded)).toEqual({
      name: '</script><script>alert("xss")</script>',
      url: 'https://docs.example.com/a&b',
    })
  })

  it('keeps only the parent and OnlyOffice origins for message checks', () => {
    expect(
      getOnlyOfficeAllowedOrigins('https://docs.example.com', 'https://app.example.com'),
    ).toEqual(['https://app.example.com', 'https://docs.example.com'])
  })

  it('requires an explicit true flag and a valid non-empty origin to enable OnlyOffice', () => {
    expect(isOnlyOfficeEnabledFromEnv(undefined, 'https://docs.example.com')).toBe(false)
    expect(isOnlyOfficeEnabledFromEnv('false', 'https://docs.example.com')).toBe(false)
    expect(isOnlyOfficeEnabledFromEnv('true', '')).toBe(false)
    expect(isOnlyOfficeEnabledFromEnv('true', 'localhost:9090')).toBe(false)
    expect(isOnlyOfficeEnabledFromEnv('true', 'https://docs.example.com')).toBe(true)
  })
})

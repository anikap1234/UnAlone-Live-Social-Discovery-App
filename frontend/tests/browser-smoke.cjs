// Run with Playwright available through npm or NODE_PATH. The local API and
// frontend must be running. Enter the development OTP from the backend terminal.
const { chromium } = require('playwright')
const fs = require('node:fs')
const path = require('node:path')
const readline = require('node:readline/promises')
const assert = require('node:assert/strict')

;(async () => {
  const output = path.resolve(__dirname, '../../output/browser-check')
  fs.mkdirSync(output, { recursive: true })
  const browser = await chromium.launch({ channel: 'chrome', headless: true, args: ['--enable-unsafe-swiftshader'] })
  const context = await browser.newContext({ viewport: { width: 1440, height: 900 },
    geolocation: { latitude: 12.9716, longitude: 77.5946 }, permissions: ['geolocation'] })
  // Automated tests do not request public OSM tiles; only normal human use does.
  await context.route('**/tile.openstreetmap.org/**', route => route.fulfill({
    contentType: 'image/png', body: Buffer.from('iVBORw0KGgoAAAANSUhEUgAAAQAAAAEACAIAAADTED8xAAADA0lEQVR4nO3VO2rEMBRA0XGY2qD9Ly9FmoDATDFMlQ24MghD7jmlQB8El7f9/H4/Fhj7OF2fx1xxHf7/mq+L++BfEABpAiBNAKQJgDQBkCYA0gRAmgBIEwBpAiBNAKQJgDQBkCYA0gRAmgBIEwBpAiBNAKQJgDQBkCYA0gRAmgBIEwBpAiBNAKQJgDQBkCYA0gRAmgBIEwBpAiBNAKQJgDQBkCYA0gRAmgBIEwBpAiBNAKQJgDQBkCYA0gRAmgBIEwBpAiBNAKQJgDQBkCYA0gRAmgBIEwBpAiBNAKRt78/r7jfAbUwA0gRAmgBIEwBpAiBNAKQJgLTnPOaKc8c+TtcXXYf/v8YEIE0ApAmANAGQJgDSBECaAEgTAGkCIE0ApAmANAGQJgDSBECaAEgTAGkCIE0ApAmANAGQJgDSBECaAEgTAGkCIE0ApAmANAGQJgDSBECaAEgTAGkCIE0ApAmANAGQJgDSBECaAEgTAGkCIE0ApAmANAGQJgDSBECaAEgTAGkCIE0ApAmANAGQJgDSBECaAEgTAGkCIE0ApAmANAGQJgDSBECaAEgTAGkCIE0ApAmANAGQJgDSBECaAEgTAGkCIE0ApAmANAGQJgDSBECaAEgTAGkCIE0ApAmANAGQJgDSBECaAEgTAGkCIE0ApAmANAGQJgDSBECaAEgTAGkCIE0ApAmANAGQJgDSBECaAEgTAGkCIE0ApAmANAGQJgDSBECaAEgTAGkCIE0ApAmANAGQJgDSBECaAEjb3p/X3W+A25gApAmANAGQJgDSBECaAEgTAGnPecwV5459nK4vug7/f40JQJoASBMAaQIgTQCkCYA0AZAmANIEQJoASBMAaQIgTQCkCYA0AZAmANIEQJoASBMAaQIgTQCkCYA0AZAmANIEQJoASBMAaQIgTQCkCYA0AZAmANIEQJoASBMAaQIgTQCkCYA0AZAmANIEQJoASBMAaQIgTQCkCYA0AZAmANIEQJoASBMAaQIgTQCkCYA0AZAmANIEQJoAeJT9Aen6IiT4X2HLAAAAAElFTkSuQmCC', 'base64')
  }))
  const page = await context.newPage()
  const errors = []
  let locationRequests = 0
  page.on('request', request => { if (request.url().endsWith('/api/location/update')) locationRequests++ })
  page.on('pageerror', error => errors.push(error.message))
  const email = 'browser-' + Date.now() + '@example.com'
  const title = 'Local MVP browser verification ' + Date.now().toString(36)
  const liveTitle = 'Live feed verification ' + Date.now().toString(36)
  const report = { email, meetupIds: [], checks: [] }
  try {
    await page.goto('http://127.0.0.1:5173')
    await page.getByRole('heading', { name: 'Start nearby' }).waitFor()
    await page.screenshot({ path: path.join(output, 'login.png'), fullPage: true })
    await page.getByLabel('Email', { exact: true }).fill(email)
    await page.getByRole('button', { name: 'Get a sign-in code' }).click()
    await page.getByLabel('Verification code').waitFor()
    console.log('OTP requested for ' + email)
    const prompt = readline.createInterface({ input: process.stdin, output: process.stdout })
    const otp = await prompt.question('Enter development OTP: ')
    prompt.close()
    await page.getByLabel('Verification code').fill(otp.trim())
    await page.getByRole('button', { name: 'Sign in', exact: true }).click()
    await page.getByRole('heading', { name: "What's happening?" }).waitFor()
    await page.getByText('Sharing while this tab is visible', { exact: true }).waitFor({ timeout: 20000 })
    report.checks.push('OTP login and mocked foreground geolocation')
    await page.getByRole('button', { name: 'Pause sharing', exact: true }).click()
    await page.getByText('Location sharing is paused', { exact: true }).waitFor()
    let before = locationRequests
    await page.waitForTimeout(8000)
    assert.equal(locationRequests, before, 'paused sharing sent location updates')
    await page.getByRole('button', { name: 'Share location', exact: true }).click()
    await page.getByText('Sharing while this tab is visible', { exact: true }).waitFor()
    await page.evaluate(() => {
      Object.defineProperty(document, 'hidden', { configurable: true, get: () => true })
      document.dispatchEvent(new Event('visibilitychange'))
    })
    await page.getByText('Sharing paused while this tab is hidden', { exact: true }).waitFor()
    before = locationRequests
    await page.waitForTimeout(8000)
    assert.equal(locationRequests, before, 'hidden tab sent location updates')
    await page.evaluate(() => {
      delete document.hidden
      document.dispatchEvent(new Event('visibilitychange'))
    })
    await page.getByText('Sharing while this tab is visible', { exact: true }).waitFor()
    report.checks.push('Pause/resume and simulated hidden-tab lifecycle stop GPS requests')
    await page.getByRole('button', { name: 'Create meetup', exact: true }).click()
    await page.getByLabel('What are you planning?').fill(title)
    await page.getByLabel('A few details').fill('Temporary verification meetup.')
    const createResponse = page.waitForResponse(r => r.url().endsWith('/api/meetups') && r.request().method() === 'POST')
    await page.getByRole('button', { name: 'Create public meetup', exact: true }).click()
    const created = (await (await createResponse).json()).meetup
    assert.ok(created?.meetupId)
    report.meetupIds.push(created.meetupId)
    await page.getByRole('heading', { name: title, exact: true }).waitFor()
    await page.screenshot({ path: path.join(output, 'discovery.png'), fullPage: true })
    report.checks.push('Meetup creation immediately appears on return to map')
    await page.getByRole('button', { name: title, exact: true }).click()
    await page.locator('.mapboxgl-popup').getByText('Temporary verification meetup.', { exact: true }).waitFor()
    await page.locator('.mapboxgl-popup-close-button').click()
    report.checks.push('Meetup map pin opens its information card')
    await page.reload()
    await page.getByRole('heading', { name: title, exact: true }).waitFor({ timeout: 20000 })
    report.checks.push('Cookie session and saved meetup survive refresh')
    await page.getByRole('button', { name: 'Activity', exact: true }).click()
    await page.locator('.connection.live').waitFor()
    await page.getByText('A quiet moment.', { exact: false }).waitFor()
    const liveResponse = await context.request.post('http://127.0.0.1:5173/api/meetups', {
      data: { title: liveTitle, description: 'Temporary event verification.', lat: 12.9716, lon: 77.5946, time: Math.floor(Date.now()/1000)+3600 }
    })
    assert.equal(liveResponse.status(), 201)
    report.meetupIds.push((await liveResponse.json()).meetup.meetupId)
    await page.getByText('New meetup: ' + liveTitle, { exact: true }).waitFor()
    report.checks.push('Feed resets on refresh and receives real WebSocket events')
    await context.setOffline(true)
    await page.locator('.connection.offline').waitFor({ timeout: 15000 })
    await context.setOffline(false)
    await page.locator('.connection.live').waitFor({ timeout: 20000 })
    await page.getByText('A quiet moment.', { exact: false }).waitFor()
    report.checks.push('Network reconnect restores the socket and resets the session feed')
    await page.getByRole('button', { name: 'Profile', exact: true }).click()
    await page.getByText(email, { exact: true }).waitFor()
    await page.getByRole('button', { name: 'Sign out', exact: true }).click()
    await page.getByRole('heading', { name: 'Start nearby' }).waitFor()
    await page.reload()
    await page.getByRole('heading', { name: 'Start nearby' }).waitFor()
    report.checks.push('Profile identity, logout, and cleared cookie session')
    await page.setViewportSize({ width: 390, height: 844 })
    await page.screenshot({ path: path.join(output, 'mobile-login.png'), fullPage: true })
    assert.ok(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth), 'mobile horizontal overflow')
    assert.deepEqual(errors, [], 'browser JavaScript errors')
    console.log(JSON.stringify(report, null, 2))
  } catch (error) {
    await page.screenshot({ path: path.join(output, 'failure.png'), fullPage: true })
    report.error = String(error)
    throw error
  } finally {
    fs.writeFileSync(path.join(output, 'report.json'), JSON.stringify(report, null, 2))
    await browser.close()
  }
})().catch(error => { console.error(error); process.exitCode = 1 })

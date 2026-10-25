import { expect, test } from 'claude-code/testing'

const PANE = {
  plugin: 'thalamus',
  component: 'Pane',
  requestId: 'thalamus-brains',
  props: {
    title: 'Thalamus brains',
    isFocused: true,
    bodyColumns: 80,
    placement: 'inline',
    scroll: { offset: 0, bodyRows: 20 },
    view: {},
  },
} as const

const LISTING = {
  gateway: '',
  brains: [
    { brain: 'coding', access: 'curate', on: true, ok: true, reason: '' },
    { brain: 'org/ref', access: 'read', on: false, ok: false, reason: 'no key' },
  ],
}

// th-hook stands behind $.process.run: brains --json answers the listing
// (a string as is), on and off answer an exit code, and each call is kept.
function fakeHook(on: any, calls: string[][], listing: object | string, exitCode = 0, opens: object[] = [], names: string[] = [], descriptions: string[] = []) {
  on('process.run', (_: unknown, e: { argv: readonly string[] }) => {
    calls.push([...e.argv.slice(1)])
    const json = e.argv[1] === 'brains'
    return { value: { exitCode: json ? 0 : exitCode, stdout: !json ? '' : typeof listing === 'string' ? listing : JSON.stringify(listing), stderr: exitCode ? 'th-hook: recall file: denied' : '' } }
  })
  on('command.register', (_: unknown, e: { name: string, description: string }) => {
    names.push(e.name)
    descriptions.push(e.description)
    return { value: { command: e.name } }
  })
  on('ui.open', (_: unknown, e: object) => {
    opens.push(e)
    return { value: { isPlaced: true } }
  })
  on('ui.focus', () => ({ value: {} }))
}

// The session registers one command, /brains, which says it shows the Cortex
// collections, and that command opens the pane.
test('the session registers /brains', async ($, on) => {
  const names: string[] = []
  const descriptions: string[] = []
  fakeHook(on, [], LISTING, 0, [], names, descriptions)
  on('session.start', (_: unknown, e: { cwd: string }) => ({ cwd: e.cwd }))
  await $.session.start({ cwd: '/tmp', surface: 'terminal', isInteractive: true })
  expect(names).toEqual(['brains'])
  expect(descriptions).toEqual(['Show the Cortex collections and turn each one on or off for the automatic recall'])
})

// The pane draws one row per brain: the cursor, a green check or a red
// cross, on or off, the access and the reason of a cross.
test('the pane shows each brain, whether it answers, and on or off', async ($, on) => {
  const calls: string[][] = []
  fakeHook(on, calls, LISTING)
  await $.command.run({ command: 'brains' })
  for (const surface of ['terminal', 'desktop'] as const) {
    const ui = await $.ui.mount({ ...PANE, surface })
    expect((await ui.find({ key: 'brain:coding' }))?.text).toMatch(/^❯ ✔ coding\s+on\s+curate$/)
    expect((await ui.find({ key: 'brain:org/ref' }))?.text).toMatch(/^\s+✘ org\/ref\s+off\s+read · no key$/)
    await ui.unmount()
  }
  expect(calls[0]).toEqual(['brains', '--json'])
})

// Enter on a brain that is on runs th-hook off, then reads the list again;
// on a brain that is off, th-hook on.
test('a press turns the brain off or on through th-hook', async ($, on) => {
  const calls: string[][] = []
  fakeHook(on, calls, LISTING)
  await $.command.run({ command: 'brains' })
  const ui = await $.ui.mount({ ...PANE, surface: 'terminal' })
  await ui.press({ key: 'brain:coding' })
  await ui.press({ key: 'brain:org/ref' })
  expect(calls).toContainEqual(['off', 'coding'])
  expect(calls).toContainEqual(['on', 'org/ref'])
  expect(calls.filter(c => c[0] === 'brains').length).toBeGreaterThanOrEqual(3)
  await ui.unmount()
})

// A gateway that does not answer, or a command that fails, is said in red,
// and the dialog grows by the rows that say it.
test('the pane says what went wrong', async ($, on) => {
  const calls: string[][] = []
  const opens: any[] = []
  fakeHook(on, calls, { gateway: 'th_brains: connection refused', brains: [] }, 1, opens)
  await $.command.run({ command: 'brains' })
  const ui = await $.ui.mount({ ...PANE, surface: 'terminal' })
  expect((await ui.find({ type: 'Text', text: /gateway/ }))?.text).toBe('✘ gateway: th_brains: connection refused')
  expect((await ui.find({ type: 'Text', text: /No brain/ }))?.text).toBe('No brain.')
  expect(opens.at(-1)).toMatchObject({ rows: 7 })
  await ui.unmount()
})

// An answer of th-hook that is not JSON is said in red, in place of the list.
test('the pane says that th-hook did not answer a list', async ($, on) => {
  const opens: any[] = []
  fakeHook(on, [], 'not json', 0, opens)
  await $.command.run({ command: 'brains' })
  const ui = await $.ui.mount({ ...PANE, surface: 'terminal' })
  expect((await ui.find({ type: 'Text', text: /JSON/ }))?.text).toMatch(/JSON/)
  expect(await ui.find({ type: 'Text', text: /No brain/ })).toBeUndefined()
  expect(opens.at(-1)).toMatchObject({ rows: 6 })
  await ui.unmount()
})

// A command that fails says why, in the pane.
test('the pane says why a change failed', async ($, on) => {
  const failed: string[][] = []
  fakeHook(on, failed, LISTING, 1)
  await $.command.run({ command: 'brains' })
  const again = await $.ui.mount({ ...PANE, surface: 'terminal' })
  await again.press({ key: 'brain:coding' })
  expect((await again.find({ type: 'Text', text: /denied/ }))?.text).toBe('th-hook: recall file: denied')
  await again.unmount()
})

// /brains opens a dialog, as /mcp does: it takes the keys, Esc closes it,
// the toasts wait, and once the list is read it is as tall as its rows.
test('the pane opens as a dialog as tall as the list', async ($, on) => {
  const opens: any[] = []
  fakeHook(on, [], LISTING, 0, opens)
  await $.command.run({ command: 'brains' })
  const ui = await $.ui.mount({ ...PANE, surface: 'terminal' })
  expect(opens[0]).toMatchObject({ id: 'thalamus-brains', focus: true, closeOnEscape: true, holdToasts: true })
  expect(opens.at(-1)).toMatchObject({ id: 'thalamus-brains', rows: LISTING.brains.length + 5 })
  await ui.unmount()
})

// The dialog reads as /mcp: a title, a count, the list, the keys at the foot.
test('the dialog has a title, a count and the keys', async ($, on) => {
  fakeHook(on, [], LISTING)
  await $.command.run({ command: 'brains' })
  const ui = await $.ui.mount({ ...PANE, surface: 'terminal' })
  expect((await ui.find({ type: 'Text', text: /^Thalamus brains$/ }))?.text).toBe('Thalamus brains')
  expect((await ui.find({ type: 'Text', text: /brains ·/ }))?.text).toBe('2 brains · automatic recall')
  expect((await ui.find({ type: 'Text', text: /navigate/ }))?.text).toBe('↑/↓ to navigate · Enter to turn on/off · Esc to cancel')
  await ui.unmount()
})

// The cursor ❯ goes to the brain that holds the focus ring. The person's
// arrow raises ui.focus with the element that takes the ring.
test('the cursor follows the focus', async ($, on) => {
  fakeHook(on, [], LISTING)
  await $.command.run({ command: 'brains' })
  const ui = await $.ui.mount({ ...PANE, surface: 'terminal' })
  await $.ui.focus({ component: 'Pane', requestId: 'thalamus-brains', element: 'brain:org/ref', origin: { kind: 'person' } } as any)
  expect((await ui.find({ key: 'brain:org/ref' }))?.text).toMatch(/^❯ ✘ org\/ref/)
  expect((await ui.find({ key: 'brain:coding' }))?.text).toMatch(/^\s+✔ coding/)
  await ui.unmount()
})

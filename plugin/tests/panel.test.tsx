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

// th-hook stands behind $.process.run: brains --json answers the listing,
// on and off answer an exit code, and each call is kept.
function fakeHook(on: any, calls: string[][], listing: object, exitCode = 0) {
  on('process.run', (_: unknown, e: { argv: readonly string[] }) => {
    calls.push([...e.argv.slice(1)])
    const json = e.argv[1] === 'brains'
    return { value: { exitCode: json ? 0 : exitCode, stdout: json ? JSON.stringify(listing) : '', stderr: exitCode ? 'th-hook: recall file: denied' : '' } }
  })
  on('command.register', () => ({ value: { command: 'th-brains' } }))
  on('ui.open', () => ({ value: { isPlaced: true } }))
}

// The pane draws one row per brain: a green check or a red cross, on or
// off, the access and the reason of a cross.
test('the pane shows each brain, whether it answers, and on or off', async ($, on) => {
  const calls: string[][] = []
  fakeHook(on, calls, LISTING)
  await $.command.run({ command: 'th-brains' })
  for (const surface of ['terminal', 'desktop'] as const) {
    const ui = await $.ui.mount({ ...PANE, surface })
    expect((await ui.find({ key: 'brain:coding' }))?.text).toMatch(/^✔ coding\s+on\s+curate$/)
    expect((await ui.find({ key: 'brain:org/ref' }))?.text).toMatch(/^✘ org\/ref\s+off\s+read · no key$/)
    await ui.unmount()
  }
  expect(calls[0]).toEqual(['brains', '--json'])
})

// Enter on a brain that is on runs th-hook off, then reads the list again;
// on a brain that is off, th-hook on.
test('a press turns the brain off or on through th-hook', async ($, on) => {
  const calls: string[][] = []
  fakeHook(on, calls, LISTING)
  await $.command.run({ command: 'th-brains' })
  const ui = await $.ui.mount({ ...PANE, surface: 'terminal' })
  await ui.press({ key: 'brain:coding' })
  await ui.press({ key: 'brain:org/ref' })
  expect(calls).toContainEqual(['off', 'coding'])
  expect(calls).toContainEqual(['on', 'org/ref'])
  expect(calls.filter(c => c[0] === 'brains').length).toBeGreaterThanOrEqual(3)
  await ui.unmount()
})

// A gateway that does not answer, or a command that fails, is said in red.
test('the pane says what went wrong', async ($, on) => {
  const calls: string[][] = []
  fakeHook(on, calls, { gateway: 'th_brains: connection refused', brains: [] }, 1)
  await $.command.run({ command: 'th-brains' })
  const ui = await $.ui.mount({ ...PANE, surface: 'terminal' })
  expect((await ui.find({ type: 'Text', text: /gateway/ }))?.text).toBe('✘ gateway: th_brains: connection refused')
  await ui.unmount()
})

// A command that fails says why, in the pane.
test('the pane says why a change failed', async ($, on) => {
  const failed: string[][] = []
  fakeHook(on, failed, LISTING, 1)
  await $.command.run({ command: 'th-brains' })
  const again = await $.ui.mount({ ...PANE, surface: 'terminal' })
  await again.press({ key: 'brain:coding' })
  expect((await again.find({ type: 'Text', text: /denied/ }))?.text).toBe('th-hook: recall file: denied')
  await again.unmount()
})

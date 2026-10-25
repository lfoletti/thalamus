import { atom, read, update } from 'claude-code'
import type { EngineInterface, Register } from 'claude-code'

import type { Listing } from '../types'

const PANE = 'thalamus-brains'
const TITLE = 'Thalamus brains'
// Change the tag when the fields of Listing change.
const listing = atom({ plugin: 'thalamus', key: 'listing' } as const,
  { gateway: '', brains: [], error: '', busy: false } as Listing, { shape: 'listing-1' })
// The key of the brain that holds the focus ring, where the cursor ❯ goes.
const cursor = atom({ plugin: 'thalamus', key: 'cursor' } as const, '')

// The lines above the list: what went wrong, or why there is no list.
function notes(l: Listing): { text: string; isError: boolean }[] {
  const out = []
  if (l.gateway !== '') out.push({ text: '✘ gateway: ' + l.gateway, isError: true })
  if (l.error !== '') out.push({ text: l.error, isError: true })
  if (l.brains.length === 0 && l.error === '') out.push({ text: l.busy ? 'Reading the brains…' : 'No brain.', isError: false })
  return out
}

// A dialog, as /mcp: it takes the keys, Esc closes it, the toasts wait. Its
// rows: the title, the count, the notes and the list between two blank rows,
// the keys.
function openDialog($: EngineInterface, l?: Listing) {
  const rows = l ? { rows: 5 + notes(l).length + l.brains.length } : {}
  return $.ui.open({ id: PANE, title: TITLE, focus: true, closeOnEscape: true, holdToasts: true, ...rows })
}

// The launcher runs th-hook (TH_HOOK_BIN, or next to it). th-hook reads and
// writes the recall file (TH_RECALL_FILE, or ~/.thalamus/recall) and holds
// every rule: the pane only draws its answers.
function hookBin($: EngineInterface): string {
  return $.plugin.root + '/bin/th-recall'
}

async function refresh($: EngineInterface): Promise<void> {
  await update($, listing, l => ({ ...l, busy: true }))
  let next: Listing
  try {
    const r = await $.process.run([hookBin($), 'brains', '--json'], { timeoutMs: 15000 })
    const v = JSON.parse(r.stdout)
    next = { gateway: v.gateway ?? '', brains: v.brains ?? [], error: '', busy: false }
  } catch (err) {
    next = { gateway: '', brains: [], error: String(err), busy: false }
  }
  await update($, listing, () => next)
  await openDialog($, next)
}

async function toggle($: EngineInterface, name: string, isOn: boolean): Promise<void> {
  await update($, listing, l => ({ ...l, busy: true }))
  const r = await $.process.run([hookBin($), isOn ? 'off' : 'on', name], { timeoutMs: 15000 })
  if (r.exitCode !== 0) {
    await update($, listing, l => ({ ...l, error: r.stderr.trim(), busy: false }))
    return
  }
  await refresh($)
}

export const register: Register = on => {
  on('session.start', async ($, e, next) => {
    await $.command.register({ name: 'brains', description: 'Show the Cortex collections and turn each one on or off for the automatic recall' })
    return next(e)
  })

  on('command.run', { command: 'brains' }, async $ => {
    await openDialog($)
    void refresh($)
    return { text: 'Thalamus brains pane opened.' }
  })

  on('ui.focus', { component: 'Pane', requestId: PANE }, async ($, e, next) => {
    await update($, cursor, () => e.element ?? '')
    return next(e)
  })

  on('ui.render', { component: 'Pane', requestId: PANE }, async ($, e) => {
    const { Box, Text, Button } = $.ui.resolve(e)
    const [l, at] = await Promise.all([read($, listing), read($, cursor)])
    const width = Math.max(...l.brains.map(b => b.brain.length), 12)
    const keys = l.brains.map(b => 'brain:' + b.brain)
    const held = keys.includes(at) ? at : keys[0]
    return (
      <Box flexDirection="column">
        <Text bold color="claude">{TITLE}</Text>
        <Text dimColor>{l.brains.length} brains · automatic recall</Text>
        <Text> </Text>
        {notes(l).map(n => <Text key={n.text} color={n.isError ? 'error' : undefined} dimColor={!n.isError}>{n.text}</Text>)}
        {l.brains.map((b, i) => (
          <Button key={keys[i]} plain autoFocus={i === 0 ? true : undefined}
            onPress={() => toggle($, b.brain, b.on)}>
            <Text color="claude">{keys[i] === held ? '❯' : ' '}</Text> <Text color={b.ok ? 'success' : 'error'}>{b.ok ? '✔' : '✘'}</Text> {b.brain.padEnd(width)}  <Text color={b.on ? 'success' : 'inactive'}>{b.on ? 'on ' : 'off'}</Text>  <Text dimColor>{b.access}{b.reason ? ' · ' + b.reason : ''}</Text>
          </Button>
        ))}
        <Text> </Text>
        <Text dimColor>↑/↓ to navigate · Enter to turn on/off · Esc to cancel{l.busy ? ' · …' : ''}</Text>
      </Box>
    )
  })
}

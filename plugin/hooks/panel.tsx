import { atom, read, update } from 'claude-code'
import type { EngineInterface, Register } from 'claude-code'

import type { Listing } from '../types'

const PANE = 'thalamus-brains'
const TITLE = 'Thalamus brains'
// Change the tag when the fields of Listing change.
const listing = atom({ plugin: 'thalamus', key: 'listing' } as const,
  { gateway: '', brains: [], error: '', busy: false } as Listing, { shape: 'listing-1' })

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
    await $.command.register({ name: 'th-brains', description: 'Show the Thalamus brains and turn each one on or off for the automatic recall' })
    return next(e)
  })

  on('command.run', { command: 'th-brains' }, async $ => {
    await $.ui.open({ id: PANE, title: TITLE, focus: true, closeOnEscape: true })
    void refresh($)
    return { text: 'Thalamus brains pane opened.' }
  })

  on('ui.render', { component: 'Pane', requestId: PANE }, async ($, e) => {
    const { Box, Text, Button } = $.ui.resolve(e)
    const l = await read($, listing)
    const width = Math.max(...l.brains.map(b => b.brain.length), 12)
    return (
      <Box flexDirection="column">
        {l.gateway !== '' && <Text color="error">✘ gateway: {l.gateway}</Text>}
        {l.error !== '' && <Text color="error">{l.error}</Text>}
        {l.brains.length === 0 && l.error === '' && <Text dimColor>{l.busy ? 'Reading the brains…' : 'No brain.'}</Text>}
        {l.brains.map((b, i) => (
          <Button key={'brain:' + b.brain} plain autoFocus={i === 0 ? true : undefined}
            onPress={() => toggle($, b.brain, b.on)}>
            <Text color={b.ok ? 'success' : 'error'}>{b.ok ? '✔' : '✘'}</Text> {b.brain.padEnd(width)}  <Text color={b.on ? 'success' : 'inactive'}>{b.on ? 'on ' : 'off'}</Text>  <Text dimColor>{b.access}{b.reason ? ' · ' + b.reason : ''}</Text>
          </Button>
        ))}
        <Text dimColor>↑↓ move · Enter on/off for the automatic recall · Esc close{l.busy ? ' · …' : ''}</Text>
      </Box>
    )
  })
}

export type Brain = { brain: string; access: string; on: boolean; ok: boolean; reason: string }
export type Listing = { gateway: string; brains: Brain[]; error: string; busy: boolean }

declare module 'claude-code' {
  interface PluginState {
    thalamus: { listing: Shaped<Listing> }
  }
}

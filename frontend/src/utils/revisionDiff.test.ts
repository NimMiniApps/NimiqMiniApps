import { describe, expect, it } from 'vitest'
import type { App, AppRevision } from '../api'
import { diffRevision } from './revisionDiff'

describe('diffRevision', () => {
  it('shows competition cycle changes to moderators', () => {
    const current = { competition_cycle: null } as App
    const revision = { competition_cycle: 1 } as AppRevision

    expect(diffRevision(current, revision)).toContainEqual({
      field: 'competition_cycle',
      label: 'Competition cycle',
      before: '—',
      after: '1',
    })
  })
})

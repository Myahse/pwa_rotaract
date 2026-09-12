import type { ClubAccessSummary } from '../api/types'

export type ClubUiProfile = 'head' | 'treasurer' | 'secretary' | 'commission_president' | 'member'

export type ManageSections = {
  members: boolean
  invites: boolean
  roles: boolean
  groups: boolean
  commissions: boolean
  accessRequests: boolean
  diary: boolean
}

export type ClubUiCaps = {
  profile: ClubUiProfile
  profileLabel: string
  nav: {
    home: boolean
    club: boolean
    messages: boolean
    cotisations: boolean
    manage: boolean
    profile: boolean
  }
  clubView: 'full' | 'groups_only' | 'hidden'
  manageSections: ManageSections
  canManageDues: boolean
  presidentCommissionIds: string[]
  defaultPath: string
}

function normalizeRoleName(name: string): string {
  return name
    .normalize('NFD')
    .replace(/[\u0300-\u036f]/g, '')
    .toLowerCase()
    .trim()
}

export function isTreasurerRole(name: string): boolean {
  const n = normalizeRoleName(name)
  return n.includes('tresorier') || n.includes('treasurer')
}

export function isSecretaryRole(name: string): boolean {
  const n = normalizeRoleName(name)
  return n.includes('secretaire') || n.includes('secretary')
}

function hasTreasurerRole(access: ClubAccessSummary): boolean {
  return access.roles.some((role) => isTreasurerRole(role.name))
}

function hasSecretaryRole(access: ClubAccessSummary): boolean {
  return access.roles.some((role) => isSecretaryRole(role.name))
}

function presidentCommissionIds(access: ClubAccessSummary): string[] {
  return (access.commissions ?? [])
    .filter((item) => item.member_role === 'president')
    .map((item) => item.commission_id)
}

export function resolveClubUiCaps(
  access: ClubAccessSummary | null,
  activeClubId: string | null,
): ClubUiCaps {
  const basePath = activeClubId ? `/clubs/${activeClubId}` : '/home'
  const fullMember: ClubUiCaps = {
    profile: 'member',
    profileLabel: 'Membre',
    nav: { home: true, club: true, messages: true, cotisations: true, manage: false, profile: true },
    clubView: 'full',
    manageSections: {
      members: false,
      invites: false,
      roles: false,
      groups: false,
      commissions: false,
      accessRequests: false,
      diary: false,
    },
    canManageDues: false,
    presidentCommissionIds: [],
    defaultPath: '/home',
  }

  if (!access || !activeClubId) return fullMember

  const presidentIds = presidentCommissionIds(access)
  const isHead = access.member_role === 'head'
  const treasurer = hasTreasurerRole(access)
  const secretary = hasSecretaryRole(access)
  const commissionPresident = presidentIds.length > 0

  if (isHead) {
    return {
      profile: 'head',
      profileLabel: 'Responsable',
      nav: { home: true, club: true, messages: true, cotisations: true, manage: true, profile: true },
      clubView: 'full',
      manageSections: {
        members: true,
        invites: true,
        roles: true,
        groups: true,
        commissions: true,
        accessRequests: true,
        diary: true,
      },
      canManageDues: true,
      presidentCommissionIds: presidentIds,
      defaultPath: '/home',
    }
  }

  if (treasurer && !secretary) {
    return {
      profile: 'treasurer',
      profileLabel: 'Trésorier',
      nav: { home: true, club: false, messages: true, cotisations: true, manage: false, profile: true },
      clubView: 'hidden',
      manageSections: {
        members: false,
        invites: false,
        roles: false,
        groups: false,
        commissions: false,
        accessRequests: false,
        diary: false,
      },
      canManageDues: true,
      presidentCommissionIds: presidentIds,
      defaultPath: `${basePath}/cotisations`,
    }
  }

  if (secretary) {
    return {
      profile: 'secretary',
      profileLabel: 'Secrétaire',
      nav: { home: true, club: false, messages: true, cotisations: false, manage: true, profile: true },
      clubView: 'hidden',
      manageSections: {
        members: true,
        invites: true,
        roles: false,
        groups: false,
        commissions: false,
        accessRequests: false,
        diary: false,
      },
      canManageDues: false,
      presidentCommissionIds: presidentIds,
      defaultPath: `${basePath}/manage`,
    }
  }

  if (commissionPresident) {
    return {
      profile: 'commission_president',
      profileLabel: 'Président de commission',
      nav: { home: true, club: true, messages: true, cotisations: false, manage: false, profile: true },
      clubView: 'groups_only',
      manageSections: {
        members: false,
        invites: false,
        roles: false,
        groups: false,
        commissions: false,
        accessRequests: false,
        diary: false,
      },
      canManageDues: false,
      presidentCommissionIds: presidentIds,
      defaultPath: basePath,
    }
  }

  return {
    ...fullMember,
    presidentCommissionIds: presidentIds,
    profileLabel: access.roles[0]?.name ?? 'Membre',
  }
}

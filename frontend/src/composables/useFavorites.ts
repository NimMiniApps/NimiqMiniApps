import { ref, watch } from 'vue'
import { getMyFavorites, addFavorite, removeFavorite } from '../api'
import { useWalletAuth } from './useWalletAuth'

const slugs = ref<Set<string>>(new Set())
let watching = false

async function loadFavorites() {
  try {
    const apps = await getMyFavorites()
    slugs.value = new Set(apps.map((a) => a.slug))
  } catch {
    slugs.value = new Set()
  }
}

export function useFavorites() {
  const { walletAddress } = useWalletAuth()

  if (!watching) {
    watching = true
    watch(
      walletAddress,
      (addr) => {
        if (addr) loadFavorites()
        else slugs.value = new Set()
      },
      { immediate: true },
    )
  }

  function isFavorite(slug: string): boolean {
    return slugs.value.has(slug)
  }

  async function toggleFavorite(slug: string) {
    const next = new Set(slugs.value)
    const wasFavorite = next.has(slug)
    if (wasFavorite) next.delete(slug)
    else next.add(slug)
    slugs.value = next
    try {
      if (wasFavorite) await removeFavorite(slug)
      else await addFavorite(slug)
    } catch {
      await loadFavorites()
    }
  }

  return { favoriteSlugs: slugs, isFavorite, toggleFavorite, loadFavorites }
}

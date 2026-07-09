export type Locale = 'en' | 'de' | 'es' | 'fr' | 'pt'

export type MessageTree = {
  nav: {
    home: string
    apps: string
    build: string
    submit: string
    developers: string
    admin: string
    more: string
  }
  theme: {
    light: string
    dark: string
  }
  footer: {
    title: string
    body: string
    curated: string
    developers: string
    githubIssues: string
  }
  common: {
    clear: string
    browseAll: string
    viewAll: string
    share: string
    copyLink: string
    copied: string
    retry: string
    all: string
    by: string
    hostedBy: string
    hostedByShort: string
    loading: string
  }
  apps: {
    title: string
    searchPlaceholder: string
    allCategories: string
    sortFeatured: string
    sortNewest: string
    sortName: string
    filteredBy: string
    filterCollection: string
    filterTag: string
    filterAsset: string
    filterDeveloper: string
    allDevelopers: string
    developerTitle: string
    developerMeta: string
    emptyTitle: string
    emptyBody: string
    emptyFilteredBody: string
    emptySearchBody: string
    errorTitle: string
    errorBody: string
    showingCount: string
    loadMore: string
  }
  collections: {
    newWeek: string
    games: string
    usdt: string
  }
  appDetail: {
    openInWallet: string
    edit: string
    editListing: string
    updatePendingHint: string
    suggestUpdate: string
    links: string
    website: string
    github: string
    about: string
    related: string
    media: string
    domain: string
    notFoundTitle: string
    notFoundBody: string
    errorTitle: string
    errorBody: string
    online: string
    offline: string
    onlineHint: string
    offlineHint: string
    offlineBanner: string
  }
  home: {
    eyebrow: string
    title: string
    subtitle: string
    searchPlaceholder: string
    searchLabel: string
    allApps: string
    browseCategories: string
    browseAll: string
    submitApp: string
    buildApp: string
    walletPrompt: string
    searchResults: string
    emptySearchTitle: string
    emptySearchBody: string
    featured: string
    newest: string
    errorTitle: string
    errorBody: string
  }
  openWallet: {
    scanTitle: string
    scanBody: string
    copyTitle: string
  }
}

const en: MessageTree = {
  nav: {
    home: 'Home',
    apps: 'Apps',
    build: 'Build',
    submit: 'Submit',
    developers: 'Developers',
    admin: 'Admin',
    more: 'More',
  },
  theme: {
    light: 'Switch to light mode',
    dark: 'Switch to dark mode',
  },
  footer: {
    title: "Don't have Nimiq Pay yet?",
    body: 'Get the free self-custodial wallet for NIM and BTC Lightning — and open every mini app here with one tap.',
    curated: 'Community-curated directory for',
    developers: 'Developers',
    githubIssues: 'GitHub issues',
  },
  common: {
    clear: 'Clear',
    browseAll: 'Browse all apps',
    viewAll: 'View all',
    share: 'Share',
    copyLink: 'Copy link',
    copied: 'Copied!',
    retry: 'Try again',
    all: 'All',
    by: 'by',
    hostedBy: 'Hosted by NimiqMiniApps.com',
    hostedByShort: 'Hosted',
    loading: 'Loading…',
  },
  apps: {
    title: 'All Apps',
    searchPlaceholder: 'Search apps…',
    allCategories: 'All categories',
    sortFeatured: 'Featured',
    sortNewest: 'Newest',
    sortName: 'Name',
    filteredBy: 'Filtered by {type}:',
    filterCollection: 'collection',
    filterTag: 'tag',
    filterAsset: 'asset',
    filterDeveloper: 'developer',
    allDevelopers: 'All developers',
    developerTitle: 'Apps by {name}',
    developerMeta: '{count} app(s) by {name}',
    emptyTitle: 'No apps here yet',
    emptyBody: 'The catalog is still growing. Browse everything or submit the next mini app for the community.',
    emptyFilteredBody: 'Nothing matches these filters. Try clearing them or pick another category.',
    emptySearchBody: 'No results for "{query}". Check the spelling or try a broader search.',
    errorTitle: "Couldn't load apps",
    errorBody: 'Something went wrong while fetching the catalog. Check your connection and try again.',
    showingCount: 'Showing {shown} of {total}',
    loadMore: 'Load more',
  },
  collections: {
    newWeek: 'New this week',
    games: 'Games',
    usdt: 'Uses USDT',
  },
  appDetail: {
    openInWallet: 'Open in Nimiq Pay',
    edit: 'Edit',
    editListing: 'Edit listing',
    updatePendingHint: 'An update is already pending review',
    suggestUpdate: 'Suggest an update',
    links: 'Links',
    website: 'Website',
    github: 'GitHub',
    about: 'About',
    related: 'Related apps',
    media: 'Screenshots & video',
    domain: 'Domain',
    notFoundTitle: 'App not found',
    notFoundBody: 'This listing may have been removed, renamed, or is still awaiting review. Head back to the catalog to discover live mini apps.',
    errorTitle: "Couldn't load this app",
    errorBody: 'Something went wrong while fetching the listing. Check your connection and try again.',
    online: 'Online',
    offline: 'Offline',
    onlineHint: 'Domain responded on the last health check',
    offlineHint: 'Domain unreachable on the last health check',
    offlineBanner: 'We could not reach this app’s domain on the last check — it may be temporarily down.',
  },
  home: {
    eyebrow: 'Community curated',
    title: 'Mini apps that live inside your Nimiq Pay wallet',
    subtitle: 'Games, maps, tools and experiments — hand-picked by the community and open with one tap, no installs, no accounts.',
    searchPlaceholder: 'Search mini apps',
    searchLabel: 'Search mini apps',
    allApps: 'All apps',
    browseCategories: 'Browse by category',
    browseAll: 'Browse all apps',
    submitApp: 'Submit your app',
    buildApp: 'Build a mini app',
    walletPrompt: 'New here? Grab the free Nimiq Pay wallet:',
    searchResults: 'Search results',
    emptySearchTitle: 'No matches',
    emptySearchBody: 'We could not find any mini apps for "{query}". Try another keyword or browse the full catalog.',
    featured: 'Featured',
    newest: 'Newest',
    errorTitle: "Couldn't load the home page",
    errorBody: 'Featured picks and categories failed to load. Refresh the page or try again in a moment.',
  },
  openWallet: {
    scanTitle: 'Scan to open',
    scanBody: 'on mobile',
    copyTitle: 'Copy Nimiq Pay open link',
  },
}

const de: MessageTree = {
  ...en,
  nav: { home: 'Start', apps: 'Apps', build: 'Entwickeln', submit: 'Einreichen', developers: 'Entwickler', admin: 'Admin', more: 'Mehr' },
  theme: { light: 'Helles Design', dark: 'Dunkles Design' },
  footer: {
    title: 'Noch kein Nimiq Pay?',
    body: 'Hol dir die kostenlose Self-Custody-Wallet für NIM und BTC Lightning — und öffne jede Mini-App hier mit einem Tipp.',
    curated: 'Community-kuratiertes Verzeichnis für',
    developers: 'Entwickler',
    githubIssues: 'GitHub-Issues',
  },
  common: { clear: 'Zurücksetzen', browseAll: 'Alle Apps', viewAll: 'Alle anzeigen', share: 'Teilen', copyLink: 'Link kopieren', copied: 'Kopiert!', retry: 'Erneut versuchen', all: 'Alle', by: 'von', hostedBy: 'Gehostet von NimiqMiniApps.com', hostedByShort: 'Gehostet', loading: 'Laden…' },
  apps: {
    title: 'Alle Apps',
    searchPlaceholder: 'Apps suchen…',
    allCategories: 'Alle Kategorien',
    sortFeatured: 'Highlights',
    sortNewest: 'Neueste',
    sortName: 'Name',
    filteredBy: 'Gefiltert nach {type}:',
    filterCollection: 'Sammlung',
    filterTag: 'Tag',
    filterAsset: 'Asset',
    filterDeveloper: 'Entwickler',
    allDevelopers: 'Alle Entwickler',
    developerTitle: 'Apps von {name}',
    developerMeta: '{count} App(s) von {name}',
    emptyTitle: 'Noch keine Apps',
    emptyBody: 'Der Katalog wächst noch. Stöbere durch alles oder reiche die nächste Mini-App für die Community ein.',
    emptyFilteredBody: 'Keine Treffer für diese Filter. Setze sie zurück oder wähle eine andere Kategorie.',
    emptySearchBody: 'Keine Ergebnisse für „{query}". Prüfe die Schreibweise oder suche breiter.',
    errorTitle: 'Apps konnten nicht geladen werden',
    errorBody: 'Beim Laden des Katalogs ist etwas schiefgelaufen. Prüfe deine Verbindung und versuche es erneut.',
    showingCount: '{shown} von {total} angezeigt',
    loadMore: 'Mehr laden',
  },
  collections: { newWeek: 'Neu diese Woche', games: 'Spiele', usdt: 'Nutzt USDT' },
  appDetail: {
    openInWallet: 'In Nimiq Pay öffnen',
    edit: 'Bearbeiten',
    editListing: 'Eintrag bearbeiten',
    updatePendingHint: 'Ein Update wartet bereits auf Prüfung',
    suggestUpdate: 'Update vorschlagen',
    links: 'Links',
    website: 'Website',
    github: 'GitHub',
    about: 'Über',
    related: 'Ähnliche Apps',
    media: 'Screenshots & Video',
    domain: 'Domain',
    notFoundTitle: 'App nicht gefunden',
    notFoundBody: 'Dieser Eintrag wurde entfernt, umbenannt oder wartet noch auf Prüfung. Entdecke live Mini-Apps im Katalog.',
    errorTitle: 'App konnte nicht geladen werden',
    errorBody: 'Beim Laden des Eintrags ist etwas schiefgelaufen. Prüfe deine Verbindung und versuche es erneut.',
    online: 'Online',
    offline: 'Offline',
    onlineHint: 'Domain hat beim letzten Health-Check geantwortet',
    offlineHint: 'Domain beim letzten Health-Check nicht erreichbar',
    offlineBanner: 'Die Domain dieser App war beim letzten Check nicht erreichbar — sie ist vielleicht vorübergehend offline.',
  },
  home: {
    eyebrow: 'Community-kuratiert',
    title: 'Mini-Apps, die in deiner Nimiq-Pay-Wallet leben',
    subtitle: 'Spiele, Karten, Tools und Experimente — von der Community ausgewählt und mit einem Tipp ohne Installation geöffnet.',
    searchPlaceholder: 'Mini-Apps suchen',
    searchLabel: 'Mini-Apps suchen',
    allApps: 'Alle Apps',
    browseCategories: 'Nach Kategorie stöbern',
    browseAll: 'Alle Apps ansehen',
    submitApp: 'App einreichen',
    buildApp: 'Mini-App entwickeln',
    walletPrompt: 'Neu hier? Hol dir die kostenlose Nimiq-Pay-Wallet:',
    searchResults: 'Suchergebnisse',
    emptySearchTitle: 'Keine Treffer',
    emptySearchBody: 'Keine Mini-Apps für „{query}" gefunden. Probiere ein anderes Stichwort oder stöbere im gesamten Katalog.',
    featured: 'Highlights',
    newest: 'Neueste',
    errorTitle: 'Startseite konnte nicht geladen werden',
    errorBody: 'Highlights und Kategorien konnten nicht geladen werden. Aktualisiere die Seite oder versuche es später erneut.',
  },
  openWallet: { scanTitle: 'Scannen zum Öffnen', scanBody: 'auf dem Handy', copyTitle: 'Nimiq-Pay-Link kopieren' },
}

const es: MessageTree = {
  ...en,
  nav: { home: 'Inicio', apps: 'Apps', build: 'Crear', submit: 'Enviar', developers: 'Desarrolladores', admin: 'Admin', more: 'Más' },
  theme: { light: 'Modo claro', dark: 'Modo oscuro' },
  footer: {
    title: '¿Aún no tienes Nimiq Pay?',
    body: 'Obtén la wallet gratuita de autocustodia para NIM y BTC Lightning — y abre cada mini app aquí con un toque.',
    curated: 'Directorio curado por la comunidad para',
    developers: 'Desarrolladores',
    githubIssues: 'Issues en GitHub',
  },
  common: { clear: 'Limpiar', browseAll: 'Ver todas las apps', viewAll: 'Ver todo', share: 'Compartir', copyLink: 'Copiar enlace', copied: '¡Copiado!', retry: 'Reintentar', all: 'Todas', by: 'por', hostedBy: 'Alojado por NimiqMiniApps.com', hostedByShort: 'Alojado', loading: 'Cargando…' },
  apps: {
    title: 'Todas las apps',
    searchPlaceholder: 'Buscar apps…',
    allCategories: 'Todas las categorías',
    sortFeatured: 'Destacadas',
    sortNewest: 'Más recientes',
    sortName: 'Nombre',
    filteredBy: 'Filtrado por {type}:',
    filterCollection: 'colección',
    filterTag: 'etiqueta',
    filterAsset: 'activo',
    filterDeveloper: 'desarrollador',
    allDevelopers: 'Todos los desarrolladores',
    developerTitle: 'Apps de {name}',
    developerMeta: '{count} app(s) de {name}',
    emptyTitle: 'Aún no hay apps',
    emptyBody: 'El catálogo sigue creciendo. Explora todo o envía la próxima mini app para la comunidad.',
    emptyFilteredBody: 'Nada coincide con estos filtros. Límpialos o elige otra categoría.',
    emptySearchBody: 'Sin resultados para «{query}». Revisa la ortografía o prueba una búsqueda más amplia.',
    errorTitle: 'No se pudieron cargar las apps',
    errorBody: 'Algo falló al obtener el catálogo. Comprueba tu conexión e inténtalo de nuevo.',
    showingCount: 'Mostrando {shown} de {total}',
    loadMore: 'Cargar más',
  },
  collections: { newWeek: 'Nuevas esta semana', games: 'Juegos', usdt: 'Usa USDT' },
  appDetail: {
    openInWallet: 'Abrir en Nimiq Pay',
    edit: 'Editar',
    editListing: 'Editar ficha',
    updatePendingHint: 'Ya hay una actualización pendiente de revisión',
    suggestUpdate: 'Sugerir cambios',
    links: 'Enlaces',
    website: 'Sitio web',
    github: 'GitHub',
    about: 'Acerca de',
    related: 'Apps relacionadas',
    media: 'Capturas y video',
    domain: 'Dominio',
    notFoundTitle: 'App no encontrada',
    notFoundBody: 'Este listado pudo haberse eliminado, renombrado o aún está en revisión. Vuelve al catálogo para descubrir mini apps activas.',
    errorTitle: 'No se pudo cargar esta app',
    errorBody: 'Algo falló al obtener el listado. Comprueba tu conexión e inténtalo de nuevo.',
    online: 'En línea',
    offline: 'Sin conexión',
    onlineHint: 'El dominio respondió en la última comprobación',
    offlineHint: 'Dominio inalcanzable en la última comprobación',
    offlineBanner: 'No pudimos alcanzar el dominio de esta app en la última comprobación — puede estar temporalmente fuera de servicio.',
  },
  home: {
    eyebrow: 'Curado por la comunidad',
    title: 'Mini apps que viven dentro de tu wallet Nimiq Pay',
    subtitle: 'Juegos, mapas, herramientas y experimentos — seleccionados por la comunidad y abiertos con un toque, sin instalaciones ni cuentas.',
    searchPlaceholder: 'Buscar mini apps',
    searchLabel: 'Buscar mini apps',
    allApps: 'Todas las apps',
    browseCategories: 'Explorar por categoría',
    browseAll: 'Ver todas las apps',
    submitApp: 'Enviar tu app',
    buildApp: 'Crear una mini app',
    walletPrompt: '¿Nuevo aquí? Descarga la wallet gratuita Nimiq Pay:',
    searchResults: 'Resultados de búsqueda',
    emptySearchTitle: 'Sin coincidencias',
    emptySearchBody: 'No encontramos mini apps para «{query}». Prueba otra palabra clave o explora el catálogo completo.',
    featured: 'Destacadas',
    newest: 'Más recientes',
    errorTitle: 'No se pudo cargar la página de inicio',
    errorBody: 'No se pudieron cargar los destacados ni las categorías. Actualiza la página o inténtalo en un momento.',
  },
  openWallet: { scanTitle: 'Escanea para abrir', scanBody: 'en móvil', copyTitle: 'Copiar enlace de Nimiq Pay' },
}

const fr: MessageTree = {
  ...en,
  nav: { home: 'Accueil', apps: 'Apps', build: 'Créer', submit: 'Soumettre', developers: 'Développeurs', admin: 'Admin', more: 'Plus' },
  theme: { light: 'Mode clair', dark: 'Mode sombre' },
  footer: {
    title: 'Pas encore Nimiq Pay ?',
    body: 'Obtenez le portefeuille gratuit en self-custody pour NIM et BTC Lightning — et ouvrez chaque mini app ici en un tap.',
    curated: 'Annuaire curaté par la communauté pour',
    developers: 'Développeurs',
    githubIssues: 'Issues GitHub',
  },
  common: { clear: 'Effacer', browseAll: 'Parcourir toutes les apps', viewAll: 'Tout voir', share: 'Partager', copyLink: 'Copier le lien', copied: 'Copié !', retry: 'Réessayer', all: 'Toutes', by: 'par', hostedBy: 'Hébergé par NimiqMiniApps.com', hostedByShort: 'Hébergé', loading: 'Chargement…' },
  apps: {
    title: 'Toutes les apps',
    searchPlaceholder: 'Rechercher des apps…',
    allCategories: 'Toutes les catégories',
    sortFeatured: 'À la une',
    sortNewest: 'Plus récentes',
    sortName: 'Nom',
    filteredBy: 'Filtré par {type} :',
    filterCollection: 'collection',
    filterTag: 'tag',
    filterAsset: 'actif',
    filterDeveloper: 'développeur',
    allDevelopers: 'Tous les développeurs',
    developerTitle: 'Apps de {name}',
    developerMeta: '{count} app(s) par {name}',
    emptyTitle: 'Aucune app ici',
    emptyBody: 'Le catalogue grandit encore. Parcourez tout ou soumettez la prochaine mini app pour la communauté.',
    emptyFilteredBody: 'Rien ne correspond à ces filtres. Effacez-les ou choisissez une autre catégorie.',
    emptySearchBody: 'Aucun résultat pour « {query} ». Vérifiez l’orthographe ou élargissez la recherche.',
    errorTitle: 'Impossible de charger les apps',
    errorBody: 'Une erreur s’est produite lors du chargement du catalogue. Vérifiez votre connexion et réessayez.',
    showingCount: '{shown} sur {total} affichées',
    loadMore: 'Charger plus',
  },
  collections: { newWeek: 'Nouveautés de la semaine', games: 'Jeux', usdt: 'Utilise USDT' },
  appDetail: {
    openInWallet: 'Ouvrir dans Nimiq Pay',
    edit: 'Modifier',
    editListing: 'Modifier la fiche',
    updatePendingHint: 'Une mise à jour est déjà en attente de validation',
    suggestUpdate: 'Proposer une modification',
    links: 'Liens',
    website: 'Site web',
    github: 'GitHub',
    about: 'À propos',
    related: 'Apps similaires',
    media: 'Captures & vidéo',
    domain: 'Domaine',
    notFoundTitle: 'App introuvable',
    notFoundBody: 'Cette fiche a peut-être été supprimée, renommée ou est encore en cours de validation. Retournez au catalogue pour découvrir les mini apps actives.',
    errorTitle: 'Impossible de charger cette app',
    errorBody: 'Une erreur s’est produite lors du chargement de la fiche. Vérifiez votre connexion et réessayez.',
    online: 'En ligne',
    offline: 'Hors ligne',
    onlineHint: 'Le domaine a répondu lors du dernier contrôle',
    offlineHint: 'Domaine injoignable lors du dernier contrôle',
    offlineBanner: 'Le domaine de cette app n’a pas répondu lors du dernier contrôle — il est peut-être temporairement indisponible.',
  },
  home: {
    eyebrow: 'Curaté par la communauté',
    title: 'Des mini apps qui vivent dans votre wallet Nimiq Pay',
    subtitle: 'Jeux, cartes, outils et expériences — sélectionnés par la communauté et ouverts en un tap, sans installation ni compte.',
    searchPlaceholder: 'Rechercher des mini apps',
    searchLabel: 'Rechercher des mini apps',
    allApps: 'Toutes les apps',
    browseCategories: 'Parcourir par catégorie',
    browseAll: 'Parcourir toutes les apps',
    submitApp: 'Soumettre votre app',
    buildApp: 'Créer une mini app',
    walletPrompt: 'Nouveau ici ? Téléchargez le wallet Nimiq Pay gratuit :',
    searchResults: 'Résultats de recherche',
    emptySearchTitle: 'Aucune correspondance',
    emptySearchBody: 'Aucune mini app pour « {query} ». Essayez un autre mot-clé ou parcourez tout le catalogue.',
    featured: 'À la une',
    newest: 'Plus récentes',
    errorTitle: 'Impossible de charger l’accueil',
    errorBody: 'Les sélections et catégories n’ont pas pu être chargées. Actualisez la page ou réessayez dans un instant.',
  },
  openWallet: { scanTitle: 'Scanner pour ouvrir', scanBody: 'sur mobile', copyTitle: 'Copier le lien Nimiq Pay' },
}

const pt: MessageTree = {
  ...en,
  nav: { home: 'Início', apps: 'Apps', build: 'Criar', submit: 'Enviar', developers: 'Desenvolvedores', admin: 'Admin', more: 'Mais' },
  theme: { light: 'Modo claro', dark: 'Modo escuro' },
  footer: {
    title: 'Ainda não tem Nimiq Pay?',
    body: 'Baixe a carteira gratuita de autocustódia para NIM e BTC Lightning — e abra cada mini app aqui com um toque.',
    curated: 'Diretório curado pela comunidade para',
    developers: 'Desenvolvedores',
    githubIssues: 'Issues no GitHub',
  },
  common: { clear: 'Limpar', browseAll: 'Ver todos os apps', viewAll: 'Ver tudo', share: 'Compartilhar', copyLink: 'Copiar link', copied: 'Copiado!', retry: 'Tentar de novo', all: 'Todos', by: 'por', hostedBy: 'Hospedado por NimiqMiniApps.com', hostedByShort: 'Hospedado', loading: 'Carregando…' },
  apps: {
    title: 'Todos os apps',
    searchPlaceholder: 'Buscar apps…',
    allCategories: 'Todas as categorias',
    sortFeatured: 'Destaques',
    sortNewest: 'Mais recentes',
    sortName: 'Nome',
    filteredBy: 'Filtrado por {type}:',
    filterCollection: 'coleção',
    filterTag: 'tag',
    filterAsset: 'ativo',
    filterDeveloper: 'desenvolvedor',
    allDevelopers: 'Todos os desenvolvedores',
    developerTitle: 'Apps de {name}',
    developerMeta: '{count} app(s) de {name}',
    emptyTitle: 'Nenhum app aqui ainda',
    emptyBody: 'O catálogo ainda está crescendo. Explore tudo ou envie o próximo mini app para a comunidade.',
    emptyFilteredBody: 'Nada corresponde a estes filtros. Limpe-os ou escolha outra categoria.',
    emptySearchBody: 'Sem resultados para «{query}». Verifique a ortografia ou tente uma busca mais ampla.',
    errorTitle: 'Não foi possível carregar os apps',
    errorBody: 'Algo deu errado ao buscar o catálogo. Verifique sua conexão e tente novamente.',
    showingCount: 'Mostrando {shown} de {total}',
    loadMore: 'Carregar mais',
  },
  collections: { newWeek: 'Novos esta semana', games: 'Jogos', usdt: 'Usa USDT' },
  appDetail: {
    openInWallet: 'Abrir no Nimiq Pay',
    edit: 'Editar',
    editListing: 'Editar ficha',
    updatePendingHint: 'Já existe uma atualização pendente de revisão',
    suggestUpdate: 'Sugerir atualização',
    links: 'Links',
    website: 'Site',
    github: 'GitHub',
    about: 'Sobre',
    related: 'Apps relacionados',
    media: 'Capturas e vídeo',
    domain: 'Domínio',
    notFoundTitle: 'App não encontrado',
    notFoundBody: 'Este item pode ter sido removido, renomeado ou ainda está em revisão. Volte ao catálogo para descobrir mini apps ativos.',
    errorTitle: 'Não foi possível carregar este app',
    errorBody: 'Algo deu errado ao buscar o item. Verifique sua conexão e tente novamente.',
    online: 'Online',
    offline: 'Offline',
    onlineHint: 'O domínio respondeu na última verificação',
    offlineHint: 'Domínio inacessível na última verificação',
    offlineBanner: 'O domínio deste app não respondeu na última verificação — pode estar temporariamente fora do ar.',
  },
  home: {
    eyebrow: 'Curado pela comunidade',
    title: 'Mini apps que vivem dentro da sua carteira Nimiq Pay',
    subtitle: 'Jogos, mapas, ferramentas e experimentos — escolhidos pela comunidade e abertos com um toque, sem instalação ou contas.',
    searchPlaceholder: 'Buscar mini apps',
    searchLabel: 'Buscar mini apps',
    allApps: 'Todos os apps',
    browseCategories: 'Explorar por categoria',
    browseAll: 'Ver todos os apps',
    submitApp: 'Enviar seu app',
    buildApp: 'Criar um mini app',
    walletPrompt: 'Novo por aqui? Baixe a carteira Nimiq Pay gratuita:',
    searchResults: 'Resultados da busca',
    emptySearchTitle: 'Nenhuma correspondência',
    emptySearchBody: 'Não encontramos mini apps para «{query}». Tente outra palavra-chave ou explore o catálogo completo.',
    featured: 'Destaques',
    newest: 'Mais recentes',
    errorTitle: 'Não foi possível carregar a página inicial',
    errorBody: 'Destaques e categorias não carregaram. Atualize a página ou tente novamente em instantes.',
  },
  openWallet: { scanTitle: 'Escaneie para abrir', scanBody: 'no celular', copyTitle: 'Copiar link do Nimiq Pay' },
}

export const messages: Record<Locale, MessageTree> = { en, de, es, fr, pt }

export const SUPPORTED_LOCALES: Locale[] = ['en', 'de', 'es', 'fr', 'pt']

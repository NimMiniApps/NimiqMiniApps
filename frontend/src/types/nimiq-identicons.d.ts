declare module '@nimiq/identicons' {
  interface IdenticonsStatic {
    svgPath: string
    placeholderToDataUrl(color: string, scale: number): string
    toDataUrl(address: string): Promise<string>
  }
  const Identicons: IdenticonsStatic
  export default Identicons
}

declare module '@nimiq/identicons/dist/identicons.min.svg?url' {
  const url: string
  export default url
}

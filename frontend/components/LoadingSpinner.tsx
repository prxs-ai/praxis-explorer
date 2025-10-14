export default function LoadingSpinner({ size = 'md' }: { size?: 'sm' | 'md' | 'lg' }) {
  const sizeClasses = {
    sm: 'w-8 h-8',
    md: 'w-16 h-16',
    lg: 'w-24 h-24'
  }

  return (
    <div className="relative inline-block">
      <div className={`${sizeClasses[size]} relative`}>
        <div className={`absolute inset-0 border-4 border-prxs-charcoal rounded-full`} />
        <div className={`absolute inset-0 border-4 border-transparent border-t-prxs-orange rounded-full animate-spin`} />
        <div className={`absolute inset-0 border-4 border-transparent border-b-prxs-cyan rounded-full animate-spin animate-reverse`} />
        <div className={`absolute inset-2 border-2 border-transparent border-l-prxs-blue rounded-full animate-spin animation-delay-200`} />
      </div>
    </div>
  )
}

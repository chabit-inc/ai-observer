import { Outlet } from 'react-router-dom'
import { Header } from './Header'
import { AppSidebar } from './Sidebar'
import { SidebarInset } from '@/components/ui/sidebar'
import { useWebSocket } from '@/hooks/useWebSocket'

export function Layout() {
  const { isConnected } = useWebSocket()

  return (
    <>
      <AppSidebar />
      <SidebarInset>
        <Header isConnected={isConnected} />
        <main className="px-6 py-6">
          <Outlet />
        </main>
      </SidebarInset>
    </>
  )
}

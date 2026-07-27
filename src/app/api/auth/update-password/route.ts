import { NextRequest, NextResponse } from 'next/server'
import { createRouteHandlerClient } from '@/lib/supabase-auth'
import { getCurrentSupabaseUser } from '@/lib/supabase-auth'
import { updatePasswordSchema } from '@/lib/schemas'
import { getJsonBody } from '@/lib/api-utils'

export async function POST(request: NextRequest) {
  try {
    const session = await getCurrentSupabaseUser()
    if (!session) {
      return NextResponse.json({ error: 'Unauthorized' }, { status: 401 })
    }

    const { password } = updatePasswordSchema.parse(await getJsonBody(request))

    const supabase = await createRouteHandlerClient(request)
    
    const { error } = await supabase.auth.updateUser({
      password,
    })

    if (error) {
      throw error
    }

    return NextResponse.json({
      message: 'Password berhasil diubah'
    })
  } catch (error: any) {
    console.error('Update password error:', error)
    
    return NextResponse.json(
      { error: 'Gagal mengubah password' },
      { status: 500 }
    )
  }
}

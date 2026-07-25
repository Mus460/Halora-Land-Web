import { prisma } from './src/lib/prisma'

async function testConnection() {
  try {
    console.log('Testing Prisma connection...')
    
    // Test 1: Simple query
    const userCount = await prisma.user.count()
    console.log('✅ Connection successful!')
    console.log(`Current user count: ${userCount}`)
    
    // Test 2: Create test user
    const testUser = await prisma.user.create({
      data: {
        namaLengkap: 'Test User',
        email: `test-${Date.now()}@example.com`,
        password: 'hashed-password-placeholder',
        role: 'USER',
        accountType: 'FREE',
        isDemo: false,
      }
    })
    
    console.log('✅ Test user created:', testUser.email)
    
    // Test 3: Query back
    const users = await prisma.user.findMany({
      take: 5,
      orderBy: { createdAt: 'desc' }
    })
    
    console.log(`✅ Found ${users.length} users`)
    
    await prisma.$disconnect()
    console.log('\n✅ All tests passed!')
    process.exit(0)
  } catch (error) {
    console.error('❌ Connection failed:', error)
    await prisma.$disconnect()
    process.exit(1)
  }
}

testConnection()

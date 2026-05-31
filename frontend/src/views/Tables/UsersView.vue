<template>
  <AdminLayout>
    <PageBreadcrumb :pageTitle="currentPageTitle" />

    <div class="space-y-6">
      <ComponentCard title="Registered RFID Cards" desc="Manage active, pending, and blocked RFID access keys.">
        
        <div class="w-full overflow-hidden">
          <div class="max-w-full overflow-x-auto custom-scrollbar">
            <table class="w-full table-auto">
              <thead>
                <tr class="border-b border-gray-200 dark:border-gray-700">
                  <th class="px-5 py-4 text-left font-medium text-gray-500 text-theme-xs dark:text-gray-400">RFID UID</th>
                  <th class="px-5 py-4 text-left font-medium text-gray-500 text-theme-xs dark:text-gray-400">User Details</th>
                  <th class="px-5 py-4 text-left font-medium text-gray-500 text-theme-xs dark:text-gray-400">Created At</th>
                  <th class="px-5 py-4 text-left font-medium text-gray-500 text-theme-xs dark:text-gray-400">Status</th>
                  <th class="px-5 py-4 text-right font-medium text-gray-500 text-theme-xs dark:text-gray-400">Actions</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-gray-200 dark:divide-gray-700">
                <tr v-for="user in users" :key="user.rfid_uid" class="hover:bg-gray-50/50 dark:hover:bg-white/[0.01] transition-colors">
                  <!-- RFID UID -->
                  <td class="px-5 py-4 whitespace-nowrap">
                    <div class="flex items-center gap-2">
                      <svg class="w-5 h-5 text-gray-400 dark:text-gray-500" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
                        <rect x="3" y="4" width="18" height="16" rx="2"></rect>
                        <circle cx="9" cy="10" r="2"></circle>
                        <line x1="14" y1="8" x2="18" y2="8"></line>
                        <line x1="14" y1="12" x2="18" y2="12"></line>
                        <line x1="6" y1="16" x2="18" y2="16"></line>
                      </svg>
                      <span class="font-mono font-bold text-gray-800 text-theme-sm dark:text-white/90">
                        {{ user.rfid_uid }}
                      </span>
                    </div>
                  </td>
                  <!-- User Details (Name & Email) -->
                  <td class="px-5 py-4 whitespace-nowrap">
                    <div class="flex flex-col">
                      <span class="font-medium text-gray-800 text-theme-sm dark:text-white/90">
                        {{ user.name || '---' }}
                      </span>
                      <span class="text-gray-500 text-theme-xs dark:text-gray-400">
                        {{ user.email || '---' }}
                      </span>
                    </div>
                  </td>
                  <!-- Created At -->
                  <td class="px-5 py-4 whitespace-nowrap text-gray-500 text-theme-sm dark:text-gray-400">
                    {{ formatDate(user.created_at) }}
                  </td>
                  <!-- Status -->
                  <td class="px-5 py-4 whitespace-nowrap">
                    <span v-if="user.is_blocked" class="rounded-full px-3 py-1 text-theme-xs font-medium inline-block bg-error-50 text-error-700 dark:bg-error-500/15 dark:text-error-400">
                      Blocked
                    </span>
                    <span v-else-if="user.is_accepted" class="rounded-full px-3 py-1 text-theme-xs font-medium inline-block bg-success-50 text-success-700 dark:bg-success-500/15 dark:text-success-400">
                      Active
                    </span>
                    <span v-else class="rounded-full px-3 py-1 text-theme-xs font-medium inline-block bg-warning-50 text-warning-700 dark:bg-warning-500/15 dark:text-orange-400">
                      Pending Activation
                    </span>
                  </td>
                  <!-- Actions -->
                  <td class="px-5 py-4 whitespace-nowrap text-right text-theme-sm">
                    <div class="flex justify-end items-center gap-2">
                      <!-- Accept Button -->
                      <button v-if="!user.is_accepted" @click="openEditModal(user, true)" class="px-3 py-1.5 text-xs font-medium rounded-md bg-brand-500 text-white hover:bg-brand-600 transition shadow-sm">
                        Accept & Edit
                      </button>
                      
                      <!-- Edit Button -->
                      <button v-if="user.is_accepted" @click="openEditModal(user, false)" class="px-3 py-1.5 text-xs font-medium rounded-md border border-gray-300 text-gray-700 hover:bg-gray-50 dark:border-gray-700 dark:text-gray-300 dark:hover:bg-white/[0.03] transition">
                        Edit
                      </button>

                      <!-- Block Button -->
                      <button v-if="user.is_accepted && !user.is_blocked" @click="toggleBlock(user)" class="px-3 py-1.5 text-xs font-medium rounded-md bg-error-50 text-error-600 hover:bg-error-100 dark:bg-error-500/10 dark:text-error-400 dark:hover:bg-error-500/20 transition">
                        Block
                      </button>

                      <!-- Unblock Button -->
                      <button v-if="user.is_blocked" @click="toggleBlock(user)" class="px-3 py-1.5 text-xs font-medium rounded-md bg-success-50 text-success-600 hover:bg-success-100 dark:bg-success-500/10 dark:text-success-400 dark:hover:bg-success-500/20 transition">
                        Unblock
                      </button>
                    </div>
                  </td>
                </tr>
                <tr v-if="users.length === 0">
                  <td colspan="5" class="px-5 py-8 text-center text-gray-500 dark:text-gray-400 text-theme-sm">
                    No RFID users registered yet. Scan a card near the reader to see it here.
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

      </ComponentCard>
    </div>

    <!-- Edit/Accept Modal -->
    <div v-if="showModal" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm">
      <div class="w-full max-w-md bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-2xl shadow-xl overflow-hidden p-6 space-y-4">
        <h3 class="text-lg font-medium text-gray-800 dark:text-white/90">
          {{ isAccepting ? 'Accept & Register RFID Card' : 'Edit User Profile' }}
        </h3>
        <p class="text-sm text-gray-500 dark:text-gray-400">
          Set the credentials for this card UID.
        </p>

        <form @submit.prevent="saveUser" class="space-y-4">
          <!-- RFID UID (readonly) -->
          <div>
            <label class="block text-xs font-medium text-gray-500 dark:text-gray-400 mb-1">RFID UID</label>
            <input type="text" :value="modalData.rfid_uid" disabled class="w-full px-3 py-2 border border-gray-200 dark:border-gray-700 bg-gray-100 dark:bg-gray-800 text-gray-500 dark:text-gray-400 rounded-lg text-sm font-mono focus:outline-none" />
          </div>

          <!-- Name -->
          <div>
            <label class="block text-xs font-medium text-gray-500 dark:text-gray-400 mb-1">Name</label>
            <input type="text" v-model="modalData.name" required placeholder="John Doe" class="w-full px-3 py-2 border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 text-gray-800 dark:text-white/90 rounded-lg text-sm focus:outline-none focus:ring-1 focus:ring-brand-500" />
          </div>

          <!-- Email -->
          <div>
            <label class="block text-xs font-medium text-gray-500 dark:text-gray-400 mb-1">E-mail</label>
            <input type="email" v-model="modalData.email" required placeholder="john.doe@example.com" class="w-full px-3 py-2 border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 text-gray-800 dark:text-white/90 rounded-lg text-sm focus:outline-none focus:ring-1 focus:ring-brand-500" />
          </div>

          <!-- Footer Buttons -->
          <div class="flex justify-end items-center gap-2 pt-2">
            <button type="button" @click="closeModal" class="px-4 py-2 text-sm font-medium border border-gray-200 dark:border-gray-700 text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-white/[0.03] rounded-lg transition">
              Cancel
            </button>
            <button type="submit" :disabled="isSaving" class="px-4 py-2 text-sm font-medium bg-brand-500 text-white hover:bg-brand-600 disabled:bg-brand-300 rounded-lg transition flex items-center gap-1">
              <svg v-if="isSaving" class="animate-spin h-4 w-4 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
              </svg>
              {{ isSaving ? 'Saving...' : (isAccepting ? 'Save & Activate' : 'Save Changes') }}
            </button>
          </div>
        </form>
      </div>
    </div>
  </AdminLayout>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { API_BASE_URL } from '@/config'
import PageBreadcrumb from "@/components/common/PageBreadcrumb.vue";
import AdminLayout from "@/components/layout/AdminLayout.vue";
import ComponentCard from "@/components/common/ComponentCard.vue";

const currentPageTitle = ref("RFID Users List");

interface User {
  rfid_uid: string;
  name: string | null;
  email: string | null;
  is_accepted: boolean;
  is_blocked: boolean;
  created_at: string;
}

const users = ref<User[]>([])
const showModal = ref(false)
const isAccepting = ref(false)
const isSaving = ref(false)

const modalData = ref({
  rfid_uid: '',
  name: '',
  email: '',
})

const fetchUsers = async () => {
  try {
    const response = await fetch(`${API_BASE_URL}/users`)
    if (!response.ok) throw new Error('Error fetching users')
    users.value = await response.json()
  } catch (err) {
    console.error("Error loading users:", err)
  }
}

const formatDate = (dateStr: string) => {
  if (!dateStr) return '---'
  const date = new Date(dateStr)
  return date.toLocaleString('pt-PT', { 
    year: 'numeric', 
    month: '2-digit', 
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}

const openEditModal = (user: User, acceptMode: boolean) => {
  isAccepting.value = acceptMode
  modalData.value = {
    rfid_uid: user.rfid_uid,
    name: user.name || '',
    email: user.email || '',
  }
  showModal.value = true
}

const closeModal = () => {
  showModal.value = false
}

const saveUser = async () => {
  isSaving.value = true
  try {
    const payload: any = {
      name: modalData.value.name,
      email: modalData.value.email,
    }
    if (isAccepting.value) {
      payload.is_accepted = true
      payload.is_blocked = false
    }

    const response = await fetch(`${API_BASE_URL}/users/${encodeURIComponent(modalData.value.rfid_uid)}`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(payload),
    })

    if (!response.ok) throw new Error('Failed to update user details')
    
    await fetchUsers()
    closeModal()
  } catch (err) {
    console.error("Error saving user:", err)
    alert("Error saving user details.")
  } finally {
    isSaving.value = false
  }
}

const toggleBlock = async (user: User) => {
  try {
    const response = await fetch(`${API_BASE_URL}/users/${encodeURIComponent(user.rfid_uid)}`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        is_blocked: !user.is_blocked,
      }),
    })

    if (!response.ok) throw new Error('Failed to toggle block status')
    await fetchUsers()
  } catch (err) {
    console.error("Error toggling block status:", err)
    alert("Error toggling block status.")
  }
}

onMounted(() => {
  fetchUsers()
})
</script>
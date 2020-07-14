import { shallowMount } from '@vue/test-utils'
import About from '@/views/About.vue'


describe('Game Page', () => {
	
	it('displays game id', () => {
		const wrapper = shallowMount(About, { propsData: {game_id: 1, valid: true, username: "Bob"} })
		expect(wrapper.find("#game_id").text()).toBe("Current Game id: 1");
	})
	
});